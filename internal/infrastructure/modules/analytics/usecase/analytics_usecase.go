package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/analytics/domain"
	"Finance-Manager-System/internal/infrastructure/modules/analytics/repository"
)

type AnalyticsUseCase struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsUseCase(repo *repository.AnalyticsRepository) *AnalyticsUseCase {
	return &AnalyticsUseCase{repo: repo}
}

func (uc *AnalyticsUseCase) resolveDates(start, end *time.Time) (time.Time, time.Time) {
	now := time.Now().UTC()
	var s, e time.Time

	if start == nil {
		s = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		s = *start
	}

	if end == nil {
		e = now
	} else {
		e = *end
	}

	return s, e
}

func (uc *AnalyticsUseCase) GetSummary(ctx context.Context, userID uuid.UUID, start, end *time.Time) (*domain.SummaryReport, error) {
	s, e := uc.resolveDates(start, end)
	return uc.repo.GetSummary(ctx, userID, s, e)
}

func (uc *AnalyticsUseCase) GetCategoryReport(ctx context.Context, userID uuid.UUID, start, end *time.Time) ([]domain.CategoryReport, error) {
	s, e := uc.resolveDates(start, end)
	return uc.repo.GetExpensesByCategory(ctx, userID, s, e)
}

func (uc *AnalyticsUseCase) GetDailyDynamics(ctx context.Context, userID uuid.UUID, start, end *time.Time, isIncome bool) ([]domain.DailyReport, error) {
	s, e := uc.resolveDates(start, end)
	return uc.repo.GetDailyDynamics(ctx, userID, s, e, isIncome)
}
