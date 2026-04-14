package usecase

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/recommendations/domain"
	"Finance-Manager-System/internal/infrastructure/modules/recommendations/repository"
)

var (
	ErrInvalidPlannedTotal = errors.New("planned_total must be greater than zero")
	ErrInvalidMonths       = errors.New("months must be between 1 and 12")
)

type RecommendationUseCase struct {
	repo *repository.RecommendationRepository
}

func NewRecommendationUseCase(repo *repository.RecommendationRepository) *RecommendationUseCase {
	return &RecommendationUseCase{repo: repo}
}

func (uc *RecommendationUseCase) GetBudgetRecommendations(
	ctx context.Context,
	userID uuid.UUID,
	plannedTotal int64,
	months int,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.BudgetRecommendation, error) {
	if plannedTotal <= 0 {
		return nil, ErrInvalidPlannedTotal
	}
	if months < 1 || months > 12 {
		return nil, ErrInvalidMonths
	}

	now := time.Now().UTC()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	windowStart := currentMonthStart.AddDate(0, -months, 0)
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)

	rows, err := uc.repo.GetMonthlyCategoryExpenses(ctx, userID, windowStart, currentMonthStart, includeHidden, accountIDs)
	if err != nil {
		return nil, err
	}

	type categoryStat struct {
		CategoryID       *uuid.UUID
		CategoryName     string
		IconURL          *string
		ShareSum         float64
		LastMonthExpense int64
	}

	monthTotals := make(map[string]int64)
	categoryByMonth := make(map[string]map[string]int64)
	categoryMeta := make(map[string]categoryStat)

	for _, row := range rows {
		monthKey := row.Month.Format("2006-01")
		catKey := "uncategorized"
		if row.CategoryID != nil {
			catKey = row.CategoryID.String()
		}

		monthTotals[monthKey] += row.Amount
		if _, ok := categoryByMonth[monthKey]; !ok {
			categoryByMonth[monthKey] = make(map[string]int64)
		}
		categoryByMonth[monthKey][catKey] = row.Amount

		if _, ok := categoryMeta[catKey]; !ok {
			categoryMeta[catKey] = categoryStat{
				CategoryID:   row.CategoryID,
				CategoryName: row.CategoryName,
				IconURL:      row.IconURL,
			}
		}

		if row.Month.Equal(lastMonthStart) {
			meta := categoryMeta[catKey]
			meta.LastMonthExpense = row.Amount
			categoryMeta[catKey] = meta
		}
	}

	for monthKey, total := range monthTotals {
		if total <= 0 {
			continue
		}
		for catKey, amount := range categoryByMonth[monthKey] {
			meta := categoryMeta[catKey]
			meta.ShareSum += float64(amount) / float64(total)
			categoryMeta[catKey] = meta
		}
	}

	totalRawShare := 0.0
	for _, stat := range categoryMeta {
		totalRawShare += stat.ShareSum / float64(months)
	}

	recommendations := make([]domain.BudgetRecommendation, 0, len(categoryMeta))
	for _, stat := range categoryMeta {
		averageShare := stat.ShareSum / float64(months)
		if averageShare <= 0 {
			continue
		}
		if totalRawShare > 0 {
			averageShare = averageShare / totalRawShare
		}
		recommendedLimit := int64(math.Round(float64(plannedTotal) * averageShare))
		overBudgetAmount := stat.LastMonthExpense - recommendedLimit
		isOverBudget := overBudgetAmount > 0
		overBudgetPercent := 0.0
		if isOverBudget && recommendedLimit > 0 {
			overBudgetPercent = (float64(overBudgetAmount) / float64(recommendedLimit)) * 100
		}

		recommendations = append(recommendations, domain.BudgetRecommendation{
			CategoryID:        stat.CategoryID,
			CategoryName:      stat.CategoryName,
			IconURL:           stat.IconURL,
			AverageShare:      averageShare,
			RecommendedLimit:  recommendedLimit,
			LastMonthExpense:  stat.LastMonthExpense,
			IsOverBudget:      isOverBudget,
			OverBudgetAmount:  maxInt64(overBudgetAmount, 0),
			OverBudgetPercent: overBudgetPercent,
		})
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].RecommendedLimit > recommendations[j].RecommendedLimit
	})

	if len(recommendations) > 0 {
		recommendedSum := int64(0)
		for i := range recommendations {
			recommendedSum += recommendations[i].RecommendedLimit
		}
		diff := plannedTotal - recommendedSum
		recommendations[0].RecommendedLimit += diff

		if recommendations[0].RecommendedLimit < 0 {
			recommendations[0].RecommendedLimit = 0
		}
	}

	return recommendations, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
