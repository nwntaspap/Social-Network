package store

import (
	"context"
	"fmt"
	"strings"

	"social-network/internal/group"
)

// SendGroupChatMessage inserts a message into the group chat room.
func (s *SQLiteStore) SendGroupChatMessage(ctx context.Context, msg *group.ChatMessage) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_chat_messages (id, group_id, sender_id, content, created_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		msg.ID, msg.GroupID, msg.SenderID, msg.Content)
	if err != nil {
		return fmt.Errorf("insert group chat message: %w", err)
	}
	return nil
}

// GetGroupChatMessages returns the most recent messages in a group chat room,
// oldest first. rowid preserves insertion order (created_at only has second
// granularity).
func (s *SQLiteStore) GetGroupChatMessages(ctx context.Context, groupID string, limit int) ([]group.ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, group_id, sender_id, content, created_at
		 FROM group_chat_messages
		 WHERE group_id = ?
		 ORDER BY rowid DESC LIMIT ?`,
		groupID, limit)
	if err != nil {
		return nil, fmt.Errorf("list group chat messages: %w", err)
	}
	defer rows.Close()

	messages := make([]group.ChatMessage, 0, limit)
	for rows.Next() {
		var m group.ChatMessage
		if err := rows.Scan(&m.ID, &m.GroupID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan group chat message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// ListGroupMemberIDs returns the IDs of every member of a group.
func (s *SQLiteStore) ListGroupMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id FROM group_members WHERE group_id = ?`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list group member ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan group member id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

// MarkGroupRead upserts the user's read marker for a group chat room. Messages
// older than the marker are considered read.
func (s *SQLiteStore) MarkGroupRead(ctx context.Context, groupID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_chat_reads (group_id, user_id, last_read_at)
		 VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(group_id, user_id) DO UPDATE SET last_read_at = CURRENT_TIMESTAMP`,
		groupID, userID)
	if err != nil {
		return fmt.Errorf("mark group chat read: %w", err)
	}
	return nil
}

// CountGroupUnread returns, per group ID, how many messages are unread for the
// user: messages sent by other members after the user's last read marker
// (or all of them if no marker exists yet).
func (s *SQLiteStore) CountGroupUnread(ctx context.Context, groupIDs []string, userID string) (map[string]int, error) {
	result := make(map[string]int, len(groupIDs))
	if len(groupIDs) == 0 {
		return result, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(groupIDs)), ",")
	args := make([]any, 0, len(groupIDs)+2)
	args = append(args, userID) // r.user_id = ?
	for _, id := range groupIDs {
		args = append(args, id)
	}
	args = append(args, userID) // m.sender_id != ?

	rows, err := s.db.QueryContext(ctx,
		`SELECT m.group_id, COUNT(*)
		 FROM group_chat_messages m
		 LEFT JOIN group_chat_reads r ON r.group_id = m.group_id AND r.user_id = ?
		 WHERE m.group_id IN (`+placeholders+`)
		   AND m.sender_id != ?
		   AND (r.last_read_at IS NULL OR datetime(m.created_at) > datetime(r.last_read_at))
		 GROUP BY m.group_id`, args...)
	if err != nil {
		return nil, fmt.Errorf("count group unread: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var groupID string
		var count int
		if err := rows.Scan(&groupID, &count); err != nil {
			return nil, fmt.Errorf("scan group unread: %w", err)
		}
		result[groupID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
