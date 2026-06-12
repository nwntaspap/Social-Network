package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arnald/forum/internal/domain/category"
	"github.com/arnald/forum/internal/domain/topic"
	"github.com/lib/pq"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) CreateCategory(ctx context.Context, category *category.Category) error {
	query := `
	INSERT INTO categories (name, description, created_by)
	VALUES ($1, $2, $3)`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
		ctx,
		category.Name,
		category.Description,
		category.CreatedBy,
	)
	if err != nil {
		if isUniqueViolation(err, "categories_name_key", "categories_name") {
			return fmt.Errorf("category with name %s already exists: %w", category.Name, ErrCategoryAlreadyExists)
		}
		if isForeignKeyViolation(err) {
			return fmt.Errorf("user with ID %s not found: %w", category.CreatedBy, ErrUserNotFound)
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *Repo) GetAllCategories(ctx context.Context, page, size int, orderBy, order, filter string) ([]category.Category, error) {
	query := `
	SELECT c.id, c.name, c.description, c.slug, c.color, c.image_path, c.created_at, c.created_by, COUNT(DISTINCT tc.topic_id) as topic_count
	FROM categories c
	LEFT JOIN topic_categories tc ON c.id = tc.category_id
	WHERE 1=1
	`
	args := make([]interface{}, 0)
	paramPos := 0

	if filter != "" {
		paramPos++
		query += fmt.Sprintf(" AND (c.name LIKE $%d OR c.description LIKE $%d)", paramPos, paramPos+1)
		paramPos++
		filterParam := "%" + filter + "%"
		args = append(args, filterParam, filterParam)
	}

	query += " GROUP BY c.id ORDER BY c." + orderBy + " " + order + fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramPos+1, paramPos+2)
	offset := (page - 1) * size
	args = append(args, size, offset)

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]category.Category, 0)
	for rows.Next() {
		var category category.Category
		err = rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Slug,
			&category.Color,
			&category.ImagePath,
			&category.CreatedAt,
			&category.CreatedBy,
			&category.TopicCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		category.Topics = make([]topic.Topic, 0)
		categories = append(categories, category)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return categories, nil
}

func (r *Repo) PopulateCategoriesWithTopics(ctx context.Context, categories []category.Category) ([]category.Category, error) {
	if len(categories) == 0 {
		return categories, nil
	}

	categoryIDs := make([]int, len(categories))
	for i, category := range categories {
		categoryIDs[i] = category.ID
	}

	placeholders := make([]string, len(categoryIDs))
	args := make([]interface{}, len(categoryIDs))
	for i, id := range categoryIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
        SELECT t.id, t.title, tc.category_id, t.created_at 
        FROM topics t
        INNER JOIN topic_categories tc ON t.id = tc.topic_id
        WHERE tc.category_id IN (`)
	queryBuilder.WriteString(strings.Join(placeholders, ","))
	queryBuilder.WriteString(") ORDER BY t.created_at DESC")
	query := queryBuilder.String()

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
	}
	defer rows.Close()

	topicsMap := make(map[int][]topic.Topic)

	for rows.Next() {
		var topic topic.Topic
		var categoryID int

		err = rows.Scan(
			&topic.ID,
			&topic.Title,
			&categoryID,
			&topic.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan topics failed: %w", err)
		}

		// Format Date
		if topic.CreatedAt != "" {
			t, parseErr := time.Parse(time.RFC3339, topic.CreatedAt)
			if parseErr == nil {
				topic.CreatedAt = t.Format("02/01/2006")
			}
		}

		topicsMap[categoryID] = append(topicsMap[categoryID], topic)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("topics rows iteration failed: %w", err)
	}

	for i := range categories {
		if topics, ok := topicsMap[categories[i].ID]; ok {
			categories[i].Topics = topics
		}
	}

	return categories, nil
}

func (r *Repo) GetTotalCategoriesCount(ctx context.Context, filter string) (int, error) {
	countQuery := `
	SELECT COUNT(*)
	FROM categories c
	WHERE 1=1
	`

	args := make([]interface{}, 0)
	if filter != "" {
		countQuery += " AND (c.name LIKE $1 OR c.description LIKE $2)"
		filterParam := "%" + filter + "%"
		args = append(args, filterParam, filterParam)
	}

	var totalCount int
	err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return totalCount, nil
}

func (r *Repo) GetCategoryByID(ctx context.Context, id int) (*category.Category, error) {
	query := `
	SELECT id, name, description, created_by, created_at
	FROM categories
	WHERE id = $1
	`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	var category category.Category
	err = stmt.QueryRowContext(ctx, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedBy,
		&category.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category with ID %d not found: %w", id, ErrCategoryNotFound)
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return &category, nil
}

func (r *Repo) DeleteCategory(ctx context.Context, id int, userID string) error {
	query := `
	DELETE FROM categories
	WHERE id = $1 AND created_by = $2
	`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("retrieving rows affected failed: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found: %w", id, ErrCategoryNotFound)
	}
	return nil
}

func (r *Repo) UpdateCategory(ctx context.Context, category *category.Category) error {
	query := `
	UPDATE categories
	SET name = $1, description = $2
	WHERE id = $3
	`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare failed: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx,
		category.Name,
		category.Description,
		category.ID,
	)
	if err != nil {
		if isUniqueViolation(err, "categories_name_key", "categories_name") {
			return fmt.Errorf("category with name %s already exists: %w", category.Name, ErrCategoryAlreadyExists)
		}
		return fmt.Errorf("exec failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("retrieving rows affected failed: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found: %w", category.ID, ErrCategoryNotFound)
	}
	return nil
}

func (r *Repo) GetAllCategorieNamesAndIDs(ctx context.Context) ([]category.Category, error) {
	query := `
	SELECT id, name, color
	FROM categories`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execture query: %w", err)
	}
	defer rows.Close()

	categories := make([]category.Category, 0)

	for rows.Next() {
		var category category.Category

		err = rows.Scan(
			&category.ID,
			&category.Name,
			&category.Color,
		)
		if err != nil {
			return nil, fmt.Errorf("scan categories failed: %w", err)
		}

		categories = append(categories, category)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("category rows iteration failed: %w", err)
	}

	return categories, nil
}

func isUniqueViolation(err error, constraints ...string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		for _, c := range constraints {
			if pqErr.Constraint == c {
				return true
			}
		}
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		return true
	}
	return false
}
