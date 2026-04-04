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

type defaultCategory struct {
	Name     string
	IsIncome bool
	IconURL  string
}

var defaultCategories = []defaultCategory{
	{Name: "Продукты", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f96b.svg"},
	{Name: "Кафе и рестораны", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f37d.svg"},
	{Name: "Транспорт", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f695.svg"},
	{Name: "Жилье", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f3e0.svg"},
	{Name: "Здоровье", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1fa7a.svg"},
	{Name: "Развлечения", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f3ad.svg"},
	{Name: "Покупки", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f6cd.svg"},
	{Name: "Подписки", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4f1.svg"},
	{Name: "Переводы", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4b8.svg"},
	{Name: "Другое", IsIncome: false, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4c2.svg"},
	{Name: "Зарплата", IsIncome: true, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4bc.svg"},
	{Name: "Кэшбэк", IsIncome: true, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4b3.svg"},
	{Name: "Проценты", IsIncome: true, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4c8.svg"},
	{Name: "Подарки", IsIncome: true, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f381.svg"},
	{Name: "Другое", IsIncome: true, IconURL: "https://raw.githubusercontent.com/twitter/twemoji/gh-pages/svg/1f4c2.svg"},
}

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) EnsureDefaultCategories(ctx context.Context, userID uuid.UUID) error {
	q := database.GetQueryer(ctx, r.db)
	query := `
		INSERT INTO Category (user_id, name_category, is_income, is_custom, icon_url)
		VALUES ($1, $2, $3, false, $4)
		ON CONFLICT (name_category, is_income, user_id) DO NOTHING
	`

	for _, cat := range defaultCategories {
		if _, err := q.ExecContext(ctx, query, userID, cat.Name, cat.IsIncome, cat.IconURL); err != nil {
			return fmt.Errorf("failed to ensure default categories: %w", err)
		}
	}

	return nil
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
