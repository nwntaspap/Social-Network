package chats

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/pkg/uuid"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		DB: db,
	}
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

// Think about  creating a chatID inside this function or in application layer.
func (r *Repo) GetOrCreateChat(ctx context.Context, userID1, userID2 string) (*chat.Chat, error) {
	lowID, highID, err := normalizeUserPair(userID1, userID2)
	if err != nil {
		return nil, err
	}

	chatID := uuid.NewProvider().NewUUID()

	// If chat already exists for this pair, INSERT is ignored because of UNIQUE constraint(user_low_id, user_high_id)
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO direct_chats (id, user_low_id, user_high_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_low_id, user_high_id) DO NOTHING
	`, chatID, lowID, highID)
	if err != nil {
		return nil, err
	}

	var c chat.Chat
	var lastMessageID sql.NullInt64
	var lastMessageAt sql.NullTime

	err = r.DB.QueryRowContext(ctx, `
		SELECT id, user_low_id, user_high_id, created_at, updated_at, last_message_id, last_message_at
		FROM direct_chats
		WHERE user_low_id = $1 AND user_high_id = $2
		LIMIT 1
	`, lowID, highID).Scan(
		&c.ID,
		&c.UserLowID,
		&c.UserHighID,
		&c.CreatedAt,
		&c.UpdatedAt,
		&lastMessageID,
		&lastMessageAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This should never happen because we just inserted or found the chat
			return nil, errors.New("chat not found after create")
		}
		return nil, err
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

func (r *Repo) GetChat(ctx context.Context, chatID string) (*chat.Chat, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	var c chat.Chat
	var lastMessageID sql.NullInt64
	var lastMessageAt sql.NullTime

	err := r.DB.QueryRowContext(ctx, `
		SELECT id, user_low_id, user_high_id, created_at, updated_at, last_message_id, last_message_at
		FROM direct_chats
		WHERE id = $1
	`, chatID).Scan(
		&c.ID,
		&c.UserLowID,
		&c.UserHighID,
		&c.CreatedAt,
		&c.UpdatedAt,
		&lastMessageID,
		&lastMessageAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("chat not found")
		}
		return nil, err
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

func (r *Repo) GetChatsForUser(ctx context.Context, userID string) ([]*chat.Chat, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT
		dc.id,
		dc.user_low_id,
		dc.user_high_id,
		dc.created_at, 
		dc.updated_at,
		dc.last_message_id,
		dc.last_message_at,
		COALESCE(cr.unread_count, 0) AS unread_count
		FROM direct_chats dc
		LEFT JOIN chat_reads cr
		ON cr.chat_id = dc.id
		AND cr.user_id = $1
		WHERE dc.user_low_id = $2 OR dc.user_high_id = $3
		ORDER BY
		CASE WHEN dc.last_message_at IS NULL THEN 1 ELSE 0 END ASC,
		dc.last_message_at DESC
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

		err := rows.Scan(
			&c.ID,
			&c.UserLowID,
			&c.UserHighID,
			&c.CreatedAt,
			&c.UpdatedAt,
			&lastMessageID,
			&lastMessageAt,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *Repo) SendMessage(ctx context.Context, chatID, senderID, content, clientMessageID string) (*chat.Message, error) {
	if chatID == "" || senderID == "" || content == "" {
		return nil, errors.New("chatID, senderID, and content cannot be empty")
	}

	now := time.Now()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert message
	var messageID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO chat_messages (chat_id, sender_id, content, created_at, client_message_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`, chatID, senderID, content, now, nullableString(clientMessageID)).Scan(&messageID)
	if err != nil {
		return nil, err
	}

	// Update direct_chats with last message info
	_, err = tx.ExecContext(ctx, `
		UPDATE direct_chats
		SET last_message_id = $1, last_message_at = $2, updated_at = $3
		WHERE id = $4
	`, messageID, now, now, chatID)
	if err != nil {
		return nil, err
	}

	// Upsert chat_reads for recipient - increment their unread count
	_, err = tx.ExecContext(ctx, `
		INSERT INTO chat_reads (chat_id, user_id, unread_count, updated_at)
		VALUES ($1, (
			SELECT CASE
				WHEN user_low_id = $2 THEN user_high_id
				ELSE user_low_id
			END
			FROM direct_chats WHERE id = $3
		), 1, $4)
		ON CONFLICT (chat_id, user_id) DO UPDATE SET 
			unread_count = chat_reads.unread_count + 1, 
			updated_at = excluded.updated_at
	`, chatID, senderID, chatID, now)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
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

func (r *Repo) GetMessagesForChat(ctx context.Context, chatID string, limit int) ([]*chat.Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, chat_id, sender_id, content, created_at, client_message_id
		FROM chat_messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		`, chatID, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanMessages(rows)
}

func (r *Repo) GetMessagesForChatBefore(ctx context.Context, chatID string, beforeMessageID int, limit int) ([]*chat.Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, chat_id, sender_id, content, created_at, client_message_id
		FROM chat_messages
		WHERE chat_id = $1 AND id < $2
		ORDER BY created_at DESC
		LIMIT $3
		`, chatID, beforeMessageID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
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

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *Repo) MarkAsRead(ctx context.Context, chatID, userID string, upToMessageID int) error {
	if chatID == "" || userID == "" {
		return errors.New("chatID and userID are required")
	}

	now := time.Now()

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO chat_reads (chat_id, user_id, last_read_message_id, last_read_at, unread_count, updated_at)
		VALUES ($1, $2, $3, $4, 0, $5)
		ON CONFLICT (chat_id, user_id) DO UPDATE SET
			last_read_message_id = excluded.last_read_message_id,
			last_read_at = excluded.last_read_at,
			unread_count = 0,
			updated_at = excluded.updated_at
			`, chatID, userID, upToMessageID, now, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetUnreadCount(ctx context.Context, chatID, userID string) (int, error) {
	if chatID == "" || userID == "" {
		return 0, errors.New("chatID and userID are required")
	}

	var count int

	err := r.DB.QueryRowContext(ctx, `
		SELECT unread_count
		FROM chat_reads
		WHERE chat_id = $1 AND user_id = $2
		`, chatID, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No row means user has never opened this chat, all messages are unread
			// but we return 0 here — let the caller decide what to do
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

func (r *Repo) GetAllUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT chat_id, unread_count
		FROM chat_reads
		WHERE user_id = $1 AND unread_count > 0
		`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	counts := make(map[string]int)

	for rows.Next() {
		var count int
		var chatID string

		err = rows.Scan(
			&chatID,
			&count,
		)
		if err != nil {
			return nil, err
		}

		counts[chatID] = count
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return counts, nil

}
