package topics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/arnald/forum/internal/domain/comment"
	"github.com/arnald/forum/internal/domain/topic"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func (r Repo) CreateTopic(ctx context.Context, topic *topic.Topic) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("transaction rollback failed: %w (original error: %w)", rollbackErr, err)
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("transaction commit failed: %w", err)
		}
	}()

	query := `
	INSERT INTO topics (user_id, title, content, image_path)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	var topicID int64
	err = stmt.QueryRowContext(
		ctx,
		topic.UserID,
		topic.Title,
		topic.Content,
		topic.ImagePath,
	).Scan(&topicID)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	categoryQuery := `
	INSERT INTO topic_categories (topic_id, category_id)
	VALUES ($1, $2)`
	categoryStmt, err := tx.PrepareContext(ctx, categoryQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare category insert: %w", err)
	}
	defer categoryStmt.Close()

	for _, categoryID := range topic.CategoryIDs {
		_, err = categoryStmt.ExecContext(ctx, topicID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to insert category %d for topic: %w", categoryID, err)
		}
	}

	return nil
}

func (r Repo) GetImagePathFromTopicID(ctx context.Context, topicID int, userID string) (imagePath string, err error) {
	query := `select image_path from topics where id=$1 and user_id=$2;`
	var imagepath sql.NullString
	err = r.DB.QueryRowContext(ctx, query, topicID, userID).Scan(&imagepath)
	if err != nil {
		return "", err
	}
	if !imagepath.Valid {
		return "", nil
	}
	return imagepath.String, nil
}
func (r Repo) UpdateTopic(ctx context.Context, topic *topic.Topic) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("transaction rollback failed: %w (original error: %w)", rollbackErr, err)
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("transaction commit failed: %w", err)
		}
	}()

	// Update topic fields
	query := `
	UPDATE topics 
	SET title = $1, content = $2, image_path = $3, updated_at = CURRENT_TIMESTAMP
	WHERE id = $4 AND user_id = $5`

	updateStmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer updateStmt.Close()

	result, err := updateStmt.ExecContext(ctx,
		topic.Title,
		topic.Content,
		topic.ImagePath,
		topic.ID,
		topic.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("topic with ID %d not found or user not authorized: %w", topic.ID, ErrTopicNotFound)
	}

	err = r.syncTopicCategories(ctx, tx, topic.ID, topic.CategoryIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) DeleteTopic(ctx context.Context, userID string, topicID int) error {
	query := `
	DELETE FROM topics
	WHERE id = $1 AND user_id = $2`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, topicID, userID)
	if err != nil {
		return fmt.Errorf("failed to execute delete statement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("topic with ID %d not found or user not authorized: %w", topicID, ErrTopicNotFound)
	}

	return nil
}

func (r Repo) GetTopicByID(ctx context.Context, topicID int, userID *string) (*topic.Topic, error) {
	query := `
	SELECT
		t.id, t.user_id, t.title, t.content, t.image_path, t.created_at, t.updated_at,
		u.username,
		STRING_AGG(DISTINCT c.id::text, ',' ORDER BY c.id::text) as category_ids,
		STRING_AGG(DISTINCT c.name, ',' ORDER BY c.name) as category_names,
		STRING_AGG(DISTINCT c.color, ',' ORDER BY c.color) as category_colors,
		COALESCE(vote_counts.upvotes, 0) as upvote_count,
		COALESCE(vote_counts.downvotes, 0) as downvote_count,
		COALESCE(vote_counts.score, 0) as vote_score`

	if userID != nil {
		query += `,
		user_vote.reaction_type as user_vote`
	}

	query += `
	FROM topics t
	LEFT JOIN users u ON t.user_id = u.id
	LEFT JOIN topic_categories tc ON t.id = tc.topic_id
	LEFT JOIN categories c ON tc.category_id = c.id
	LEFT JOIN (
		SELECT
			topic_id,
			COUNT(CASE WHEN reaction_type = 1 THEN 1 END) as upvotes,
			COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as downvotes,
			(COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END)) as score
			FROM votes
			WHERE comment_id IS NULL
			GROUP BY topic_id
	) vote_counts ON t.id = vote_counts.topic_id`

	args := make([]interface{}, 0)
	paramPos := 0

	if userID != nil {
		paramPos++
		query += fmt.Sprintf(`
		LEFT JOIN votes user_vote ON t.id = user_vote.topic_id
			AND user_vote.user_id = $%d
			AND user_vote.comment_id IS NULL`, paramPos)
	}

	paramPos++
	query += fmt.Sprintf(` WHERE t.id = $%d`, paramPos)
	query += ` GROUP BY t.id, t.user_id, t.title, t.content, t.image_path, t.created_at, t.updated_at, u.username, vote_counts.upvotes, vote_counts.downvotes, vote_counts.score`

	if userID != nil {
		query += `, user_vote.reaction_type`
	}

	if userID != nil {
		args = append(args, *userID)
	}

	args = append(args, topicID)

	var topicResult topic.Topic
	var userVote sql.NullInt32
	var categoryIDs, categoryNames, categoryColors sql.NullString

	scanFields := []interface{}{
		&topicResult.ID,
		&topicResult.UserID,
		&topicResult.Title,
		&topicResult.Content,
		&topicResult.ImagePath,
		&topicResult.CreatedAt,
		&topicResult.UpdatedAt,
		&topicResult.OwnerUsername,
		&categoryIDs,
		&categoryNames,
		&categoryColors,
		&topicResult.UpvoteCount,
		&topicResult.DownvoteCount,
		&topicResult.VoteScore,
	}

	if userID != nil {
		scanFields = append(scanFields, &userVote)
	}

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, args...).Scan(scanFields...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("topic with ID %d not found: %w", topicID, ErrTopicNotFound)
		}
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}

	parseCategoryData(&topicResult, categoryIDs, categoryNames, categoryColors)

	// Format Dates
	if topicResult.CreatedAt != "" {
		t, parseErr := time.Parse(time.RFC3339, topicResult.CreatedAt)
		if parseErr == nil {
			topicResult.CreatedAt = t.Format("02/01/2006")
		}
	}

	if topicResult.UpdatedAt != "" {
		t, parseErr := time.Parse(time.RFC3339, topicResult.UpdatedAt)
		if parseErr == nil {
			topicResult.UpdatedAt = t.Format("02/01/2006")
		}
	}

	if userID != nil && userVote.Valid {
		vote := int(userVote.Int32)
		topicResult.UserVote = &vote
	}

	return &topicResult, nil
}

func (r Repo) GetTotalTopicsCount(ctx context.Context, filter string, categoryID int) (int, error) {
	countQuery := `
    SELECT COUNT(DISTINCT t.id) 
    FROM topics t`

	args := make([]interface{}, 0)
	paramPos := 0

	// Add junction table join only if filtering by category
	if categoryID > 0 {
		countQuery += `
        LEFT JOIN topic_categories tc ON t.id = tc.topic_id`
	}

	countQuery += `
    WHERE 1=1`

	if filter != "" {
		paramPos++
		countQuery += fmt.Sprintf(" AND (t.title LIKE $%d OR t.content LIKE $%d)", paramPos, paramPos+1)
		paramPos++
		filterParam := "%" + filter + "%"
		args = append(args, filterParam, filterParam)
	}

	if categoryID > 0 {
		paramPos++
		countQuery += fmt.Sprintf(" AND tc.category_id = $%d", paramPos)
		args = append(args, categoryID)
	}

	var totalCount int
	err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return totalCount, nil
}

func (r Repo) GetAllTopics(ctx context.Context, page, size, categoryID int, orderBy, order, filter string, userID *string) ([]topic.Topic, error) {
	query := `
    SELECT 
        t.id, t.user_id, t.title, t.content, t.image_path, t.created_at, t.updated_at,
        u.username,
        STRING_AGG(DISTINCT c.id::text, ',' ORDER BY c.id::text) as category_ids,
        STRING_AGG(DISTINCT c.name, ',' ORDER BY c.name) as category_names,
        STRING_AGG(DISTINCT c.color, ',' ORDER BY c.color) as category_colors,
        COALESCE(vote_counts.upvotes, 0) as upvote_count,
        COALESCE(vote_counts.downvotes, 0) as downvote_count,
        COALESCE(vote_counts.score, 0) as vote_score`

	if userID != nil {
		query += `,
        user_votes.reaction_type as user_vote`
	}

	query += `
    FROM topics t
    LEFT JOIN users u ON t.user_id = u.id
    LEFT JOIN topic_categories tc ON t.id = tc.topic_id
    LEFT JOIN categories c ON tc.category_id = c.id
    LEFT JOIN (
        SELECT
            topic_id,
            COUNT(CASE WHEN reaction_type = 1 THEN 1 END) as upvotes,
            COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as downvotes,
            (COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END)) as score
            FROM votes
            WHERE comment_id IS NULL
            GROUP BY topic_id
        ) vote_counts ON t.id = vote_counts.topic_id`

	args := make([]interface{}, 0)
	paramPos := 0

	if userID != nil {
		paramPos++
		query += fmt.Sprintf(`
        LEFT JOIN votes user_votes ON t.id = user_votes.topic_id
            AND user_votes.user_id = $%d
            AND user_votes.comment_id IS NULL`, paramPos)
	}

	query += ` WHERE 1=1`

	if userID != nil {
		args = append(args, *userID)
	}

	if filter != "" {
		paramPos++
		query += fmt.Sprintf(" AND (t.title LIKE $%d OR t.content LIKE $%d)", paramPos, paramPos+1)
		paramPos++
		filterParam := "%" + filter + "%"
		args = append(args, filterParam, filterParam)
	}

	if categoryID > 0 {
		paramPos++
		query += fmt.Sprintf(" AND t.id IN (SELECT topic_id FROM topic_categories WHERE category_id = $%d)", paramPos)
		args = append(args, categoryID)
	}

	// GROUP BY is essential when using STRING_AGG
	query += " GROUP BY t.id, t.user_id, t.title, t.content, t.image_path, t.created_at, t.updated_at, u.username, vote_counts.upvotes, vote_counts.downvotes, vote_counts.score"

	if userID != nil {
		query += ", user_votes.reaction_type"
	}

	orderByClause := "t." + orderBy

	if orderBy == "vote_score" {
		orderByClause = "vote_counts.score"
	}

	paramPos++
	query += " ORDER BY " + orderByClause + " " + order + fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramPos, paramPos+1)
	offset := (page - 1) * size
	args = append(args, size, offset)

	fmt.Println(query)
	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
	}
	defer rows.Close()

	topics := make([]topic.Topic, 0)
	for rows.Next() {
		var topic topic.Topic
		var userVote sql.NullInt32
		var categoryIDs, categoryNames, categoryColors sql.NullString

		scanFields := []interface{}{
			&topic.ID,
			&topic.UserID,
			&topic.Title,
			&topic.Content,
			&topic.ImagePath,
			&topic.CreatedAt,
			&topic.UpdatedAt,
			&topic.OwnerUsername,
			&categoryIDs,
			&categoryNames,
			&categoryColors,
			&topic.UpvoteCount,
			&topic.DownvoteCount,
			&topic.VoteScore,
		}

		if userID != nil {
			scanFields = append(scanFields, &userVote)
		}

		err = rows.Scan(scanFields...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		parseCategoryData(&topic, categoryIDs, categoryNames, categoryColors)

		// Format dates
		if topic.CreatedAt != "" {
			t, parseErr := time.Parse(time.RFC3339, topic.CreatedAt)
			if parseErr == nil {
				topic.CreatedAt = t.Format("02/01/2006")
			}
		}

		if topic.UpdatedAt != "" {
			t, parseErr := time.Parse(time.RFC3339, topic.UpdatedAt)
			if parseErr == nil {
				topic.UpdatedAt = t.Format("02/01/2006")
			}
		}

		if userID != nil && userVote.Valid {
			vote := int(userVote.Int32)
			topic.UserVote = &vote
		}

		topic.Comments = make([]comment.Comment, 0)
		topics = append(topics, topic)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return topics, nil
}

func parseCategoryData(t *topic.Topic, categoryIDs, categoryNames, categoryColors sql.NullString) {
	if !categoryIDs.Valid || categoryIDs.String == "" {
		return
	}

	ids := strings.Split(categoryIDs.String, ",")
	t.CategoryIDs = make([]int, 0, len(ids))
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr != "" {
			id, parseErr := strconv.Atoi(idStr)
			if parseErr == nil {
				t.CategoryIDs = append(t.CategoryIDs, id)
			}
		}
	}

	if categoryNames.Valid && categoryNames.String != "" {
		t.CategoryNames = strings.Split(categoryNames.String, ",")
		for i := range t.CategoryNames {
			t.CategoryNames[i] = strings.TrimSpace(t.CategoryNames[i])
		}
	}

	if categoryColors.Valid && categoryColors.String != "" {
		t.CategoryColors = strings.Split(categoryColors.String, ",")
		for i := range t.CategoryColors {
			t.CategoryColors[i] = strings.TrimSpace(t.CategoryColors[i])
		}
	}
}

// syncTopicCategories handles all category synchronization logic.
func (r Repo) syncTopicCategories(ctx context.Context, tx *sql.Tx, topicID int, newCategoryIDs []int) error {
	// Get existing categories
	existingCategoryIDs := make([]int, 0)
	rows, err := tx.QueryContext(ctx,
		"SELECT category_id FROM topic_categories WHERE topic_id = $1",
		topicID)
	if err != nil {
		return fmt.Errorf("failed to get existing categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var catID int
		scanErr := rows.Scan(&catID)
		if scanErr != nil {
			return fmt.Errorf("failed to scan category: %w", scanErr)
		}
		existingCategoryIDs = append(existingCategoryIDs, catID)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return fmt.Errorf("error iterating categories: %w", rowsErr)
	}

	// Build maps for comparison
	existingMap := make(map[int]bool)
	for _, id := range existingCategoryIDs {
		existingMap[id] = true
	}

	newMap := make(map[int]bool)
	for _, id := range newCategoryIDs {
		newMap[id] = true
	}

	// Delete removed categories
	for _, catID := range existingCategoryIDs {
		if !newMap[catID] {
			_, err = tx.ExecContext(ctx,
				"DELETE FROM topic_categories WHERE topic_id = $1 AND category_id = $2",
				topicID, catID)
			if err != nil {
				return fmt.Errorf("failed to delete category %d: %w", catID, err)
			}
		}
	}

	// Insert new categories
	if len(newCategoryIDs) > 0 {
		insertStmt, err := tx.PrepareContext(ctx,
			"INSERT INTO topic_categories (topic_id, category_id) VALUES ($1, $2)")
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer insertStmt.Close()

		for _, catID := range newCategoryIDs {
			if !existingMap[catID] {
				_, err := insertStmt.ExecContext(ctx, topicID, catID)
				if err != nil {
					return fmt.Errorf("failed to insert category %d: %w", catID, err)
				}
			}
		}
	}

	return nil
}
