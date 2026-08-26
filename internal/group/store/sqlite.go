package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"social-network/internal/group"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) CreateGroup(ctx context.Context, g *group.Group) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO groups (id, title, description, creator_id) VALUES (?, ?, ?, ?)`,
		g.ID, g.Title, g.Description, g.CreatorID)
	if err != nil {
		return err
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT created_at FROM groups WHERE id = ?`, g.ID).Scan(&g.CreatedAt)
	return err
}

func (s *SQLiteStore) AddMember(ctx context.Context, groupID, userID string, role group.Role) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_members (group_id, user_id, role) VALUES (?, ?, ?)`,
		groupID, userID, string(role))
	return err
}

func (s *SQLiteStore) GetGroupByID(ctx context.Context, groupID string) (*group.Group, error) {
	var g group.Group
	var updatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, description, creator_id, created_at, updated_at
		 FROM groups WHERE id = ?`, groupID).Scan(
		&g.ID, &g.Title, &g.Description, &g.CreatorID, &g.CreatedAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrGroupNotFound
		}
		return nil, fmt.Errorf("get group: %w", err)
	}
	g.UpdatedAt = database.ResolveTime(updatedAt, g.CreatedAt)
	return &g, nil
}

func (s *SQLiteStore) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)`,
		groupID, userID).Scan(&exists)
	return exists, err
}

func (s *SQLiteStore) GetMemberRole(ctx context.Context, groupID, userID string) (group.Role, error) {
	var role string
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", group.ErrNotMember
		}
		return "", fmt.Errorf("get member role: %w", err)
	}
	return group.Role(role), nil
}

func (s *SQLiteStore) GetGroupAdmins(ctx context.Context, groupID string) ([]string, error) {
	var users []string
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id FROM group_members WHERE group_id = ? AND role in (?,?)`,
		groupID, group.RoleAdmin, group.RoleCreator)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	return users, rows.Err()
}

func (s *SQLiteStore) CreateInvitation(ctx context.Context, inv *group.Invitation) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_invitations (id, group_id, inviter_id, invitee_id) VALUES (?, ?, ?, ?)`,
		inv.ID, inv.GroupID, inv.InviterID, inv.InviteeID)
	return err
}

func (s *SQLiteStore) DeleteInvitation(ctx context.Context, groupID, inviteeID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM group_invitations WHERE group_id = ? AND invitee_id = ?`,
		groupID, inviteeID)
	return err
}

func (s *SQLiteStore) GetInvitation(ctx context.Context, groupID, inviteeID string) (*group.Invitation, error) {
	var inv group.Invitation
	err := s.db.QueryRowContext(ctx,
		`SELECT id, group_id, inviter_id, invitee_id, created_at
		 FROM group_invitations WHERE group_id = ? AND invitee_id = ?`,
		groupID, inviteeID).Scan(
		&inv.ID, &inv.GroupID, &inv.InviterID, &inv.InviteeID, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrInvitationNotFound
		}
		return nil, fmt.Errorf("get invitation: %w", err)
	}
	return &inv, nil
}

func (s *SQLiteStore) CreateJoinRequest(ctx context.Context, jr *group.JoinRequest) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_join_requests (id, group_id, requester_id) VALUES (?, ?, ?)`,
		jr.ID, jr.GroupID, jr.RequesterID)
	return err
}

func (s *SQLiteStore) DeleteJoinRequest(ctx context.Context, groupID, requesterID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM group_join_requests WHERE group_id = ? AND requester_id = ?`,
		groupID, requesterID)
	return err
}

func (s *SQLiteStore) GetJoinRequest(ctx context.Context, groupID, requesterID string) (*group.JoinRequest, error) {
	var jr group.JoinRequest
	err := s.db.QueryRowContext(ctx,
		`SELECT id, group_id, requester_id, created_at
		 FROM group_join_requests WHERE group_id = ? AND requester_id = ?`,
		groupID, requesterID).Scan(
		&jr.ID, &jr.GroupID, &jr.RequesterID, &jr.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrJoinRequestNotFound
		}
		return nil, fmt.Errorf("get join request: %w", err)
	}
	return &jr, nil
}

func (s *SQLiteStore) IsInvited(ctx context.Context, groupID, inviteeID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM group_invitations WHERE group_id = ? AND invitee_id = ?)`,
		groupID, inviteeID).Scan(&exists)
	return exists, err
}

func (s *SQLiteStore) HasPendingRequest(ctx context.Context, groupID, requesterID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM group_join_requests WHERE group_id = ? AND requester_id = ?)`,
		groupID, requesterID).Scan(&exists)
	return exists, err
}

func (s *SQLiteStore) CreatePost(ctx context.Context, p *group.Post) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_posts (id, group_id, author_id, title, content, image_path) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.GroupID, p.AuthorID, p.Title, p.Content, p.ImagePath)
	return err
}

func (s *SQLiteStore) GetPostByID(ctx context.Context, postID string) (*group.Post, error) {
	var p group.Post
	var imagePath sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, group_id, author_id, title, content, image_path, created_at
		 FROM group_posts WHERE id = ?`, postID).Scan(
		&p.ID, &p.GroupID, &p.AuthorID, &p.Title, &p.Content, &imagePath, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrPostNotFound
		}
		return nil, fmt.Errorf("get group post: %w", err)
	}
	p.ImagePath = imagePath.String
	return &p, nil
}

func (s *SQLiteStore) GetPostsByGroupID(ctx context.Context, groupID, userID string, page, size int) ([]group.Post, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_posts WHERE group_id = ?`, groupID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count group posts: %w", err)
	}

	offset := (page - 1) * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.group_id, p.author_id, p.title, p.content, p.image_path, p.created_at, p.updated_at,
		        COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0),
		        COALESCE(cc.comments_count, 0),
		        uv.reaction_type
		 FROM group_posts p
		 LEFT JOIN (
		     SELECT post_id,
		            COUNT(CASE WHEN reaction_type = 1 THEN 1 END) AS upvotes,
		            COUNT(CASE WHEN reaction_type = -1 THEN 1 END) AS downvotes,
		            COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END) AS score
		     FROM group_post_votes
		     GROUP BY post_id
		 ) vc ON p.id = vc.post_id
		 LEFT JOIN (
		     SELECT post_id, COUNT(*) AS comments_count
		     FROM group_post_comments
		     GROUP BY post_id
		 ) cc ON p.id = cc.post_id
		 LEFT JOIN group_post_votes uv ON p.id = uv.post_id AND uv.user_id = ?
		 WHERE p.group_id = ?
		 ORDER BY p.created_at DESC LIMIT ? OFFSET ?`,
		userID, groupID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list group posts: %w", err)
	}
	defer rows.Close()

	var posts []group.Post
	for rows.Next() {
		var p group.Post
		var imagePath sql.NullString
		var updatedAt sql.NullTime
		var userVote sql.NullInt32
		if err := rows.Scan(&p.ID, &p.GroupID, &p.AuthorID, &p.Title, &p.Content, &imagePath,
			&p.CreatedAt, &updatedAt, &p.UpvoteCount, &p.DownvoteCount, &p.VoteScore,
			&p.CommentsCount, &userVote); err != nil {
			return nil, 0, fmt.Errorf("scan group post: %w", err)
		}
		p.ImagePath = imagePath.String
		p.UpdatedAt = database.ResolveTime(updatedAt, p.CreatedAt)
		if userVote.Valid {
			v := int(userVote.Int32)
			p.UserVote = &v
		}
		posts = append(posts, p)
	}
	return posts, total, rows.Err()
}

func (s *SQLiteStore) ListGroups(ctx context.Context, page, size int) ([]group.Group, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count groups: %w", err)
	}

	offset := (page - 1) * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, description, creator_id, created_at, updated_at
		 FROM groups ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var groups []group.Group
	for rows.Next() {
		var g group.Group
		var updatedAt sql.NullTime
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.CreatorID, &g.CreatedAt, &updatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan group: %w", err)
		}
		g.UpdatedAt = database.ResolveTime(updatedAt, g.CreatedAt)
		groups = append(groups, g)
	}
	return groups, total, rows.Err()
}

func groupSearchFilter(query string) (string, []any) {
	if query == "" {
		return "", nil
	}
	pattern := "%" + query + "%"
	where := `WHERE title LIKE ? OR description LIKE ?`
	args := []any{pattern, pattern}
	return where, args
}

func (s *SQLiteStore) SearchGroups(ctx context.Context, query string, page, size int) ([]group.Group, int, error) {
	where, args := groupSearchFilter(query)

	var total int
	countArgs := args
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups `+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count groups: %w", err)
	}

	offset := (page - 1) * size
	args = append(args, size, offset)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, description, creator_id, created_at, updated_at
		 FROM groups `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search groups: %w", err)
	}
	defer rows.Close()

	var groups []group.Group
	for rows.Next() {
		var g group.Group
		var updatedAt sql.NullTime
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.CreatorID, &g.CreatedAt, &updatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan group: %w", err)
		}
		g.UpdatedAt = database.ResolveTime(updatedAt, g.CreatedAt)
		groups = append(groups, g)
	}
	return groups, total, rows.Err()
}

// ListUserGroups returns the groups a user belongs to, newest first.
func (s *SQLiteStore) ListUserGroups(ctx context.Context, userID string, page, size int) ([]group.Group, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*)
		 FROM groups g JOIN group_members gm ON gm.group_id = g.id
		 WHERE gm.user_id = ?`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count user groups: %w", err)
	}

	offset := (page - 1) * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT g.id, g.title, g.description, g.creator_id, g.created_at, g.updated_at
		 FROM groups g JOIN group_members gm ON gm.group_id = g.id
		 WHERE gm.user_id = ?
		 ORDER BY g.created_at DESC LIMIT ? OFFSET ?`,
		userID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list user groups: %w", err)
	}
	defer rows.Close()

	var groups []group.Group
	for rows.Next() {
		var g group.Group
		var updatedAt sql.NullTime
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.CreatorID, &g.CreatedAt, &updatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user group: %w", err)
		}
		g.UpdatedAt = database.ResolveTime(updatedAt, g.CreatedAt)
		groups = append(groups, g)
	}
	return groups, total, rows.Err()
}

func (s *SQLiteStore) GetPendingInvitations(ctx context.Context, userID string) ([]group.Invitation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, group_id, inviter_id, invitee_id, created_at
		 FROM group_invitations WHERE invitee_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []group.Invitation
	for rows.Next() {
		var inv group.Invitation
		if err := rows.Scan(&inv.ID, &inv.GroupID, &inv.InviterID, &inv.InviteeID, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

func (s *SQLiteStore) GetSentInvitationInviteeIDs(ctx context.Context, groupID, inviterID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT invitee_id FROM group_invitations WHERE group_id = ? AND inviter_id = ?`,
		groupID, inviterID)
	if err != nil {
		return nil, fmt.Errorf("list sent invitation invitee ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan invitee id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *SQLiteStore) GetPendingJoinRequests(ctx context.Context, groupID string) ([]group.JoinRequest, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, group_id, requester_id, created_at
		 FROM group_join_requests WHERE group_id = ? ORDER BY created_at ASC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	defer rows.Close()

	var requests []group.JoinRequest
	for rows.Next() {
		var jr group.JoinRequest
		if err := rows.Scan(&jr.ID, &jr.GroupID, &jr.RequesterID, &jr.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan join request: %w", err)
		}
		requests = append(requests, jr)
	}
	return requests, rows.Err()
}

func (s *SQLiteStore) CountMembers(ctx context.Context, groupID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_members WHERE group_id = ?`, groupID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return count, nil
}

func (s *SQLiteStore) GetGroupMembers(ctx context.Context, groupID string, page, size int) ([]group.Member, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_members WHERE group_id = ?`, groupID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count members: %w", err)
	}

	offset := (page - 1) * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT group_id, user_id, role, joined_at
		 FROM group_members WHERE group_id = ?
		 ORDER BY joined_at ASC LIMIT ? OFFSET ?`,
		groupID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []group.Member
	for rows.Next() {
		var m group.Member
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, 0, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, total, rows.Err()
}

func (s *SQLiteStore) RemoveMember(ctx context.Context, groupID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID)
	return err
}

func (s *SQLiteStore) UpdateGroup(ctx context.Context, id, title, description string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE groups SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		title, description, id)
	return err
}

func (s *SQLiteStore) DeleteGroup(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM group_post_comments WHERE post_id IN (SELECT id FROM group_posts WHERE group_id = ?)`, id); err != nil {
		return fmt.Errorf("delete post comments: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM group_posts WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("delete posts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM group_members WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("delete members: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM group_join_requests WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("delete join requests: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM group_invitations WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("delete invitations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetJoinRequestByID(ctx context.Context, id string) (*group.JoinRequest, error) {
	var jr group.JoinRequest
	err := s.db.QueryRowContext(ctx,
		`SELECT id, group_id, requester_id, created_at
		 FROM group_join_requests WHERE id = ?`, id).Scan(
		&jr.ID, &jr.GroupID, &jr.RequesterID, &jr.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrJoinRequestNotFound
		}
		return nil, fmt.Errorf("get join request by id: %w", err)
	}
	return &jr, nil
}
