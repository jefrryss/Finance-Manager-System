package repository

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
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

func (r *AnalyticsRepository) GetSummary(
	ctx context.Context,
	userID uuid.UUID,
	start, end time.Time,
	includeHidden bool,
	accountIDs []uuid.UUID,
) (*domain.SummaryReport, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN is_income = true THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN is_income = false THEN amount ELSE 0 END), 0) AS total_expense
		FROM Transactions
		WHERE user_id = $1 AND completed_at >= $2 AND completed_at <= $3
	`

	args := []interface{}{userID, start, end}
	nextArg := 4
	if !includeHidden {
		query += " AND is_hidden = false"
	}
	if len(accountIDs) > 0 {
		placeholders := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			placeholders[i] = fmt.Sprintf("$%d", nextArg)
			args = append(args, id)
			nextArg++
		}
		query += " AND account_id IN (" + strings.Join(placeholders, ", ") + ")"
	}

	var report domain.SummaryReport
	err := r.db.GetContext(ctx, &report, query, args...)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *AnalyticsRepository) GetByCategory(
	ctx context.Context,
	userID uuid.UUID,
	start, end time.Time,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.CategoryReport, error) {
	query := `
		SELECT
			t.category_id,
			COALESCE(c.name_category, 'Без категории') AS category_name,
			c.icon_url AS icon_url,
			COALESCE(SUM(t.amount), 0) AS total_amount,
			COALESCE(
				SUM(t.amount) * 100.0 / NULLIF(SUM(SUM(t.amount)) OVER (), 0),
				0
			) AS share_percent
		FROM Transactions t
		LEFT JOIN Category c ON c.category_id = t.category_id
		WHERE t.user_id = $1 AND t.is_income = $2 AND t.completed_at >= $3 AND t.completed_at <= $4
	`

	args := []interface{}{userID, isIncome, start, end}
	nextArg := 5
	if !includeHidden {
		query += " AND t.is_hidden = false"
	}
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
		GROUP BY t.category_id, c.name_category, c.icon_url
		ORDER BY total_amount DESC
	`

	reports := make([]domain.CategoryReport, 0)
	err := r.db.SelectContext(ctx, &reports, query, args...)
	return reports, err
}

func (r *AnalyticsRepository) GetDailyDynamics(
	ctx context.Context,
	userID uuid.UUID,
	start, end time.Time,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.DailyReport, error) {
	query := `
		SELECT completed_at::DATE AS date, COALESCE(SUM(amount), 0) AS total_amount
		FROM Transactions
		WHERE user_id = $1 AND is_income = $2 AND completed_at >= $3 AND completed_at <= $4
	`

	args := []interface{}{userID, isIncome, start, end}
	nextArg := 5
	if !includeHidden {
		query += " AND is_hidden = false"
	}
	if len(accountIDs) > 0 {
		placeholders := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			placeholders[i] = fmt.Sprintf("$%d", nextArg)
			args = append(args, id)
			nextArg++
		}
		query += " AND account_id IN (" + strings.Join(placeholders, ", ") + ")"
	}

	query += `
		GROUP BY completed_at::DATE
		ORDER BY date ASC
	`

	reports := make([]domain.DailyReport, 0)
	err := r.db.SelectContext(ctx, &reports, query, args...)
	return reports, err
}

func (r *AnalyticsRepository) GetMonthlyDynamics(
	ctx context.Context,
	userID uuid.UUID,
	start, end time.Time,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.MonthlyReport, error) {
	query := `
		SELECT date_trunc('month', completed_at)::DATE AS month, COALESCE(SUM(amount), 0) AS total_amount
		FROM Transactions
		WHERE user_id = $1 AND is_income = $2 AND completed_at >= $3 AND completed_at <= $4
	`

	args := []interface{}{userID, isIncome, start, end}
	nextArg := 5
	if !includeHidden {
		query += " AND is_hidden = false"
	}
	if len(accountIDs) > 0 {
		placeholders := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			placeholders[i] = fmt.Sprintf("$%d", nextArg)
			args = append(args, id)
			nextArg++
		}
		query += " AND account_id IN (" + strings.Join(placeholders, ", ") + ")"
	}

	query += `
		GROUP BY date_trunc('month', completed_at)::DATE
		ORDER BY month ASC
	`

	reports := make([]domain.MonthlyReport, 0)
	err := r.db.SelectContext(ctx, &reports, query, args...)
	return reports, err
}

func (r *AnalyticsRepository) CompareCategoryPeriods(
	ctx context.Context,
	userID uuid.UUID,
	firstStart, firstEnd time.Time,
	secondStart, secondEnd time.Time,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.CategoryCompareReport, error) {
	firstPeriodRows, err := r.GetByCategory(ctx, userID, firstStart, firstEnd, isIncome, includeHidden, accountIDs)
	if err != nil {
		return nil, err
	}
	secondPeriodRows, err := r.GetByCategory(ctx, userID, secondStart, secondEnd, isIncome, includeHidden, accountIDs)
	if err != nil {
		return nil, err
	}

	type aggregate struct {
		categoryID   *uuid.UUID
		categoryName string
		iconURL      *string
		firstAmount  int64
		secondAmount int64
	}

	aggregates := make(map[string]*aggregate)
	keyFor := func(categoryID *uuid.UUID, categoryName string) string {
		if categoryID != nil {
			return categoryID.String()
		}
		return "uncategorized:" + categoryName
	}

	for _, row := range firstPeriodRows {
		key := keyFor(row.CategoryID, row.CategoryName)
		aggregates[key] = &aggregate{
			categoryID:   row.CategoryID,
			categoryName: row.CategoryName,
			iconURL:      row.IconURL,
			firstAmount:  row.TotalAmount,
		}
	}

	for _, row := range secondPeriodRows {
		key := keyFor(row.CategoryID, row.CategoryName)
		if existing, ok := aggregates[key]; ok {
			existing.secondAmount = row.TotalAmount
			if existing.iconURL == nil && row.IconURL != nil {
				existing.iconURL = row.IconURL
			}
			if existing.categoryName == "" {
				existing.categoryName = row.CategoryName
			}
			continue
		}
		aggregates[key] = &aggregate{
			categoryID:   row.CategoryID,
			categoryName: row.CategoryName,
			iconURL:      row.IconURL,
			secondAmount: row.TotalAmount,
		}
	}

	result := make([]domain.CategoryCompareReport, 0, len(aggregates))
	for _, row := range aggregates {
		delta := row.secondAmount - row.firstAmount
		var deltaPercent *float64
		if row.firstAmount != 0 {
			value := float64(delta) * 100.0 / float64(row.firstAmount)
			deltaPercent = &value
		}
		result = append(result, domain.CategoryCompareReport{
			CategoryID:         row.categoryID,
			CategoryName:       row.categoryName,
			IconURL:            row.iconURL,
			FirstPeriodAmount:  row.firstAmount,
			SecondPeriodAmount: row.secondAmount,
			DeltaAmount:        delta,
			DeltaPercent:       deltaPercent,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		absI := math.Abs(float64(result[i].DeltaAmount))
		absJ := math.Abs(float64(result[j].DeltaAmount))
		if absI == absJ {
			return result[i].CategoryName < result[j].CategoryName
		}
		return absI > absJ
	})

	return result, nil
}
