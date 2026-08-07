package store

import (
	"context"
	"fmt"

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
