package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/category/domain"
)

type CategoryRepo struct {
	db *sqlx.DB
}

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) AddCategory(ctx context.Context, category *domain.Category) (uuid.UUID, error) {
	q := database.GetQueryer(ctx, r.db)

	query := `
        INSERT INTO Category (user_id, name_category, is_income, is_custom, icon_url) 
        VALUES (:user_id, :name_category, :is_income, :is_custom, :icon_url)
        RETURNING category_id
    `

	queryStr, args, err := sqlx.Named(query, category)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to process named query: %w", err)
	}

	queryStr = q.Rebind(queryStr)

	var generatedID uuid.UUID
	err = q.QueryRowContext(ctx, queryStr, args...).Scan(&generatedID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add category: %w", err)
	}

	return generatedID, nil
}

func (r *CategoryRepo) GetCategoriesByUser(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	q := database.GetQueryer(ctx, r.db)
	categories := make([]domain.Category, 0)
	query := `SELECT * FROM Category WHERE user_id = $1 ORDER BY name_category ASC`

	err := q.SelectContext(ctx, &categories, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepo) GetCategoryByID(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) (*domain.Category, error) {
	q := database.GetQueryer(ctx, r.db)
	var cat domain.Category
	query := `SELECT * FROM Category WHERE user_id = $1 AND category_id = $2`

	err := q.GetContext(ctx, &cat, query, userID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &cat, nil
}

func (r *CategoryRepo) UpdateCategory(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID, newName string, newIconURL *string) error {
	q := database.GetQueryer(ctx, r.db)
	query := `
        UPDATE Category 
        SET name_category = $1, icon_url = $2 
        WHERE category_id = $3 AND user_id = $4
    `

	result, err := q.ExecContext(ctx, query, newName, newIconURL, categoryID, userID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryRepo) DeleteCategory(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) error {
	q := database.GetQueryer(ctx, r.db)
	query := `DELETE FROM Category WHERE user_id = $1 AND category_id = $2`

	result, err := q.ExecContext(ctx, query, userID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
