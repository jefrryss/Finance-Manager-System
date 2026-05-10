package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Finance-Manager-System/internal/infrastructure/modules/recommendations/domain"
)

type RecommendationRepository struct {
	db *sqlx.DB
}

func NewRecommendationRepository(db *sqlx.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

func (r *RecommendationRepository) GetMonthlyCategoryExpenses(
	ctx context.Context,
	userID uuid.UUID,
	start time.Time,
	end time.Time,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.MonthlyCategoryExpense, error) {
	query := `
		SELECT
			date_trunc('month', t.completed_at)::date AS month,
			t.category_id,
			COALESCE(c.name_category, 'Без категории') AS category_name,
			c.icon_url,
			COALESCE(SUM(t.amount), 0) AS amount
		FROM Transactions t
		LEFT JOIN Category c ON c.category_id = t.category_id
		WHERE t.user_id = $1
		  AND t.is_income = false
		  AND t.completed_at >= $2
		  AND t.completed_at < $3
		  AND ($4 OR t.is_hidden = false)
	`

	args := []interface{}{userID, start, end, includeHidden}
	nextArg := 5

	if len(accountIDs) > 0 {
		placeholders := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			placeholders[i] = fmt.Sprintf("$%d", nextArg)
			args = append(args, id)
			nextArg++
		}
		query += " AND t.account_id IN (" + strings.Join(placeholders, ", ") + ")"
	}

	query += `
		GROUP BY date_trunc('month', t.completed_at)::date, t.category_id, c.name_category, c.icon_url
		ORDER BY month ASC, amount DESC
	`

	rows := make([]domain.MonthlyCategoryExpense, 0)
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	return rows, nil
}
