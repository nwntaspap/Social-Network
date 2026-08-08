package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"social-network/internal/chat"
	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func normalizeUserPair(a, b string) (string, string, error) {
	if a == "" || b == "" {
		return "", "", errors.New("user IDs cannot be empty")
	}
	if a == b {
		return "", "", errors.New("cannot create chat with self")
	}
	if a < b {
		return a, b, nil
	}
	return b, a, nil
}

// GetOrCreateChat returns an existing chat between two users or creates a new one.
func (s *SQLiteStore) GetOrCreateChat(ctx context.Context, userOneID, userTwoID string) (*chat.Chat, error) {
	lowID, highID, err := normalizeUserPair(userOneID, userTwoID)
	if err != nil {
		return nil, err
	}

	chatID := uuid.NewProvider().NewUUID()

	_, err = s.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO chats (id, user_one_id, user_two_id)
		VALUES (?, ?, ?)
	`, chatID, lowID, highID)
	if err != nil {
		return nil, err
	}

	return s.getChatByPair(ctx, lowID, highID)
}

func (s *SQLiteStore) getChatByPair(ctx context.Context, userOneID, userTwoID string) (*chat.Chat, error) {
	var c chat.Chat
	var lastMessageID sql.NullInt64
	var lastMessageAt sql.NullTime
	var updatedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_one_id, user_two_id, created_at, updated_at, last_message_id, last_message_at
		FROM chats
		WHERE user_one_id = ? AND user_two_id = ?
		LIMIT 1
	`, userOneID, userTwoID).Scan(
		&c.ID,
		&c.UserOneID,
		&c.UserTwoID,
		&c.CreatedAt,
		&updatedAt,
		&lastMessageID,
		&lastMessageAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("chat not found after create")
		}
		return nil, err
	}

	c.UpdatedAt = updatedAt.Time
	if !updatedAt.Valid {
		c.UpdatedAt = c.CreatedAt
	}

	if lastMessageID.Valid {
		id := int(lastMessageID.Int64)
		c.LastMessageID = &id
	}
	if lastMessageAt.Valid {
		t := lastMessageAt.Time
		c.LastMessageAt = &t
	}

	return &c, nil
}

// GetChat returns a single chat by ID.
func (s *SQLiteStore) GetChat(ctx context.Context, chatID string) (*chat.Chat, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	var c chat.Chat
	var lastMessageID sql.NullInt64
	var lastMessageAt sql.NullTime
	var updatedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_one_id, user_two_id, created_at, updated_at, last_message_id, last_message_at
		FROM chats
		WHERE id = ?
	`, chatID).Scan(
		&c.ID,
		&c.UserOneID,
		&c.UserTwoID,
		&c.CreatedAt,
		&updatedAt,
		&lastMessageID,
		&lastMessageAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, chat.ErrChatNotFound
		}
		return nil, err
	}

	c.UpdatedAt = updatedAt.Time
	if !updatedAt.Valid {
		c.UpdatedAt = c.CreatedAt
	}

	if lastMessageID.Valid {
		id := int(lastMessageID.Int64)
		c.LastMessageID = &id
	}
	if lastMessageAt.Valid {
		t := lastMessageAt.Time
		c.LastMessageAt = &t
	}

	return &c, nil
}

// GetChatsForUser returns all chats for a user with unread counts, ordered by last message.
func (s *SQLiteStore) GetChatsForUser(ctx context.Context, userID string) ([]*chat.Chat, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			c.id,
			c.user_one_id,
			c.user_two_id,
			c.created_at,
			c.updated_at,
			c.last_message_id,
			c.last_message_at,
			COALESCE(cr.unread_count, 0) AS unread_count
		FROM chats c
		LEFT JOIN chat_reads cr
			ON cr.chat_id = c.id
			AND cr.user_id = ?
		WHERE c.user_one_id = ? OR c.user_two_id = ?
		ORDER BY
			CASE WHEN c.last_message_at IS NULL THEN 1 ELSE 0 END ASC,
			c.last_message_at DESC
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*chat.Chat

	for rows.Next() {
		var c chat.Chat
		var lastMessageID sql.NullInt64
		var lastMessageAt sql.NullTime
		var updatedAt sql.NullTime

		err = rows.Scan(
			&c.ID,
			&c.UserOneID,
			&c.UserTwoID,
			&c.CreatedAt,
			&updatedAt,
			&lastMessageID,
			&lastMessageAt,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
		}

		c.UpdatedAt = updatedAt.Time
		if !updatedAt.Valid {
			c.UpdatedAt = c.CreatedAt
		}

		if lastMessageID.Valid {
			id := int(lastMessageID.Int64)
			c.LastMessageID = &id
		}
		if lastMessageAt.Valid {
			t := lastMessageAt.Time
			c.LastMessageAt = &t
		}

		chats = append(chats, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

// SendMessage inserts a new message and updates the chat's last message fields.
func (s *SQLiteStore) SendMessage(ctx context.Context, chatID, senderID, content, clientMessageID string) (*chat.Message, error) {
	if chatID == "" || senderID == "" || content == "" {
		return nil, errors.New("chatID, senderID, and content cannot be empty")
	}

	now := time.Now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO messages (chat_id, sender_id, content, created_at, client_message_id)
		VALUES (?, ?, ?, ?, ?)
	`, chatID, senderID, content, now, nullableString(clientMessageID))
	if err != nil {
		return nil, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE chats
		SET last_message_id = ?, last_message_at = ?, updated_at = ?
		WHERE id = ?
	`, messageID, now, now, chatID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO chat_reads (chat_id, user_id, unread_count, updated_at)
		VALUES (?, (
			SELECT CASE
				WHEN user_one_id = ? THEN user_two_id
				ELSE user_one_id
			END
			FROM chats WHERE id = ?
		), 1, ?)
		ON CONFLICT (chat_id, user_id) DO UPDATE SET
			unread_count = unread_count + 1,
			updated_at = excluded.updated_at
	`, chatID, senderID, chatID, now)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	m := &chat.Message{
		ID:        int(messageID),
		ChatID:    chatID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: now,
	}
	if clientMessageID != "" {
		m.ClientMessageID = &clientMessageID
	}

	return m, nil
}

// GetMessagesForChat returns the most recent messages in a chat, oldest first.
// created_at only has second granularity, so id breaks ties in insertion order.
func (s *SQLiteStore) GetMessagesForChat(ctx context.Context, chatID string, limit int) ([]*chat.Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, chat_id, sender_id, content, created_at, client_message_id
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}
	return reverseMessages(messages), nil
}

// GetMessagesForChatBefore returns messages before a given ID for pagination,
// oldest first. The window is the most recent messages older than beforeID.
func (s *SQLiteStore) GetMessagesForChatBefore(ctx context.Context, chatID string, beforeID int, limit int) ([]*chat.Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, chat_id, sender_id, content, created_at, client_message_id
		FROM messages
		WHERE chat_id = ? AND id < ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, chatID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}
	return reverseMessages(messages), nil
}

func reverseMessages(messages []*chat.Message) []*chat.Message {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages
}

func scanMessages(rows *sql.Rows) ([]*chat.Message, error) {
	messages := make([]*chat.Message, 0)

	for rows.Next() {
		var m chat.Message
		var clientMessageID sql.NullString

		err := rows.Scan(
			&m.ID,
			&m.ChatID,
			&m.SenderID,
			&m.Content,
			&m.CreatedAt,
			&clientMessageID,
		)
		if err != nil {
			return nil, err
		}

		if clientMessageID.Valid {
			m.ClientMessageID = &clientMessageID.String
		}

		messages = append(messages, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkAsRead records that a user has read up to a given message.
func (s *SQLiteStore) MarkAsRead(ctx context.Context, chatID, userID string, upToMessageID int) error {
	if chatID == "" || userID == "" {
		return errors.New("chatID and userID are required")
	}

	now := time.Now()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_reads (chat_id, user_id, last_read_message_id, last_read_at, unread_count, updated_at)
		VALUES (?, ?, ?, ?, 0, ?)
		ON CONFLICT (chat_id, user_id) DO UPDATE SET
			last_read_message_id = excluded.last_read_message_id,
			last_read_at = excluded.last_read_at,
			unread_count = 0,
			updated_at = excluded.updated_at
	`, chatID, userID, upToMessageID, now, now)
	return err
}

// GetUnreadCount returns the number of unread messages for a user in a chat.
func (s *SQLiteStore) GetUnreadCount(ctx context.Context, chatID, userID string) (int, error) {
	if chatID == "" || userID == "" {
		return 0, errors.New("chatID and userID are required")
	}

	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT unread_count
		FROM chat_reads
		WHERE chat_id = ? AND user_id = ?
	`, chatID, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

// GetAllUnreadCounts returns unread counts per chat for a user.
func (s *SQLiteStore) GetAllUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT chat_id, unread_count
		FROM chat_reads
		WHERE user_id = ? AND unread_count > 0
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var count int
		var chatID string
		if err = rows.Scan(&chatID, &count); err != nil {
			return nil, err
		}
		counts[chatID] = count
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
