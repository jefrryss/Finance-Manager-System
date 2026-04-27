package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/analytics/domain"
	"Finance-Manager-System/internal/infrastructure/modules/analytics/repository"
)

type AnalyticsUseCase struct {
	repo *repository.AnalyticsRepository
}

var ErrInvalidPeriod = errors.New("invalid period")

func NewAnalyticsUseCase(repo *repository.AnalyticsRepository) *AnalyticsUseCase {
	return &AnalyticsUseCase{repo: repo}
}

func (uc *AnalyticsUseCase) resolveDates(start, end *time.Time, period string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	var s, e time.Time

	if start != nil {
		s = *start
	} else {
		switch period {
		case "", "month":
			s = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		case "week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := now.AddDate(0, 0, -(weekday - 1))
			s = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, time.UTC)
		case "day":
			s = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		default:
			return time.Time{}, time.Time{}, ErrInvalidPeriod
		}
	}

	if end != nil {
		e = *end
	} else {
		e = now
	}

	if e.Before(s) {
		return time.Time{}, time.Time{}, ErrInvalidPeriod
	}

	return s, e, nil
}

func (uc *AnalyticsUseCase) GetSummary(
	ctx context.Context,
	userID uuid.UUID,
	start, end *time.Time,
	period string,
	includeHidden bool,
	accountIDs []uuid.UUID,
) (*domain.SummaryReport, error) {
	s, e, err := uc.resolveDates(start, end, period)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetSummary(ctx, userID, s, e, includeHidden, accountIDs)
}

func (uc *AnalyticsUseCase) GetCategoryReport(
	ctx context.Context,
	userID uuid.UUID,
	start, end *time.Time,
	period string,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.CategoryReport, error) {
	s, e, err := uc.resolveDates(start, end, period)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetByCategory(ctx, userID, s, e, isIncome, includeHidden, accountIDs)
}

func (uc *AnalyticsUseCase) GetDailyDynamics(
	ctx context.Context,
	userID uuid.UUID,
	start, end *time.Time,
	period string,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.DailyReport, error) {
	s, e, err := uc.resolveDates(start, end, period)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetDailyDynamics(ctx, userID, s, e, isIncome, includeHidden, accountIDs)
}

func (uc *AnalyticsUseCase) GetMonthlyDynamics(
	ctx context.Context,
	userID uuid.UUID,
	start, end *time.Time,
	period string,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.MonthlyReport, error) {
	s, e, err := uc.resolveDates(start, end, period)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetMonthlyDynamics(ctx, userID, s, e, isIncome, includeHidden, accountIDs)
}

func monthRange(month time.Time) (time.Time, time.Time) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}

func (uc *AnalyticsUseCase) CompareCategoriesByMonths(
	ctx context.Context,
	userID uuid.UUID,
	firstMonth, secondMonth time.Time,
	isIncome bool,
	includeHidden bool,
	accountIDs []uuid.UUID,
) ([]domain.CategoryCompareReport, error) {
	firstStart, firstEnd := monthRange(firstMonth)
	secondStart, secondEnd := monthRange(secondMonth)
	return uc.repo.CompareCategoryPeriods(
		ctx,
		userID,
		firstStart,
		firstEnd,
		secondStart,
		secondEnd,
		isIncome,
		includeHidden,
		accountIDs,
	)
}
