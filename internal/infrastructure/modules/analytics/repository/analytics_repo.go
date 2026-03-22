package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Finance-Manager-System/internal/infrastructure/modules/analytics/domain"
)

type AnalyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.SummaryReport, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN is_income = true THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN is_income = false THEN amount ELSE 0 END), 0) AS total_expense
		FROM Transactions
		WHERE user_id = $1 AND is_hidden = false AND completed_at >= $2 AND completed_at <= $3
	`

	var report domain.SummaryReport
	err := r.db.GetContext(ctx, &report, query, userID, start, end)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *AnalyticsRepository) GetExpensesByCategory(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.CategoryReport, error) {
	query := `
		SELECT category_id, COALESCE(SUM(amount), 0) AS total_amount
		FROM Transactions
		WHERE user_id = $1 AND is_income = false AND is_hidden = false AND completed_at >= $2 AND completed_at <= $3
		GROUP BY category_id
		ORDER BY total_amount DESC
	`

	reports := make([]domain.CategoryReport, 0)
	err := r.db.SelectContext(ctx, &reports, query, userID, start, end)
	return reports, err
}

func (r *AnalyticsRepository) GetDailyDynamics(ctx context.Context, userID uuid.UUID, start, end time.Time, isIncome bool) ([]domain.DailyReport, error) {
	query := `
		SELECT completed_at::DATE AS date, COALESCE(SUM(amount), 0) AS total_amount
		FROM Transactions
		WHERE user_id = $1 AND is_income = $2 AND is_hidden = false AND completed_at >= $3 AND completed_at <= $4
		GROUP BY completed_at::DATE
		ORDER BY date ASC
	`

	reports := make([]domain.DailyReport, 0)
	err := r.db.SelectContext(ctx, &reports, query, userID, isIncome, start, end)
	return reports, err
}
