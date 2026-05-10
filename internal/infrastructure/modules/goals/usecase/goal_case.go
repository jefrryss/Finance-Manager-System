package usecase

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/goals/domain"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type GoalRepository interface {
	AddGoal(ctx context.Context, goal *domain.Goal) (uuid.UUID, error)
	GetGoalsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetGoalByID(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) (*domain.Goal, error)
	UpdateGoal(ctx context.Context, goal *domain.Goal) error
	DeleteGoal(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) error
	AddContribution(ctx context.Context, contribution *domain.GoalContribution) (uuid.UUID, error)
	IncreaseCurrentAmount(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, amount int64) error
	GetGoalContributions(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) ([]domain.GoalContribution, error)
	GetMonthlyContributions(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, start time.Time, end time.Time) ([]domain.MonthlyContribution, error)
}

type GoalUseCase struct {
	repo      GoalRepository
	transRepo GoalTransactionRepository
	txManager database.TxManager
}

type GoalTransactionRepository interface {
	GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*transactionDomain.Transaction, error)
}

func NewGoalUseCase(repo GoalRepository, transRepo GoalTransactionRepository, txManager database.TxManager) *GoalUseCase {
	return &GoalUseCase{repo: repo, transRepo: transRepo, txManager: txManager}
}

func (uc *GoalUseCase) CreateGoal(ctx context.Context, userID uuid.UUID, name string, targetAmount int64, targetDate *time.Time) (uuid.UUID, error) {
	goal, err := domain.NewGoal(userID, name, targetAmount, targetDate)
	if err != nil {
		return uuid.Nil, err
	}
	return uc.repo.AddGoal(ctx, goal)
}

func (uc *GoalUseCase) GetGoals(ctx context.Context, userID uuid.UUID) ([]domain.GoalSummary, error) {
	goals, err := uc.repo.GetGoalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	summaries := make([]domain.GoalSummary, 0, len(goals))
	for _, g := range goals {
		summaries = append(summaries, buildSummary(g, now))
	}
	return summaries, nil
}

func (uc *GoalUseCase) GetGoalDetails(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) (*domain.GoalDetails, error) {
	goal, err := uc.repo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return nil, err
	}

	contributions, err := uc.repo.GetGoalContributions(ctx, userID, goalID)
	if err != nil {
		return nil, err
	}

	forecast, err := uc.buildForecast(ctx, userID, goalID, goal.CurrentAmount, goal.TargetAmount, 3)
	if err != nil {
		return nil, err
	}

	excessAmount := int64(0)
	redirectSuggestions := make([]domain.GoalRedirectSuggestion, 0)
	if goal.CurrentAmount > goal.TargetAmount {
		excessAmount = goal.CurrentAmount - goal.TargetAmount
		goals, goalsErr := uc.repo.GetGoalsByUser(ctx, userID)
		if goalsErr != nil {
			return nil, goalsErr
		}

		for _, candidate := range goals {
			if candidate.GoalID == goal.GoalID {
				continue
			}
			if candidate.CurrentAmount >= candidate.TargetAmount {
				continue
			}
			needed := candidate.TargetAmount - candidate.CurrentAmount
			redirectSuggestions = append(redirectSuggestions, domain.GoalRedirectSuggestion{
				GoalID:       candidate.GoalID,
				NameGoal:     candidate.NameGoal,
				NeededAmount: needed,
			})
		}

		sort.Slice(redirectSuggestions, func(i, j int) bool {
			return redirectSuggestions[i].NeededAmount < redirectSuggestions[j].NeededAmount
		})
	}

	return &domain.GoalDetails{
		Summary:             buildSummary(*goal, time.Now().UTC()),
		Contributions:       contributions,
		Forecast:            forecast,
		ExcessAmount:        excessAmount,
		RedirectSuggestions: redirectSuggestions,
	}, nil
}

func (uc *GoalUseCase) UpdateGoal(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, name string, targetAmount int64, targetDate *time.Time) error {
	existingGoal, err := uc.repo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return err
	}

	validated, err := domain.NewGoal(userID, name, targetAmount, targetDate)
	if err != nil {
		return err
	}

	existingGoal.NameGoal = validated.NameGoal
	existingGoal.TargetAmount = validated.TargetAmount
	existingGoal.TargetDate = validated.TargetDate
	existingGoal.UpdatedAt = time.Now().UTC()

	return uc.repo.UpdateGoal(ctx, existingGoal)
}

func (uc *GoalUseCase) DeleteGoal(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) error {
	return uc.repo.DeleteGoal(ctx, userID, goalID)
}

func (uc *GoalUseCase) AddContribution(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, amount int64, contributionDate *time.Time, transactionID *uuid.UUID) (uuid.UUID, error) {
	if _, err := uc.repo.GetGoalByID(ctx, userID, goalID); err != nil {
		return uuid.Nil, err
	}

	if transactionID != nil {
		if uc.transRepo == nil {
			return uuid.Nil, domain.ErrGoalNotFound
		}
		trans, err := uc.transRepo.GetTransaction(ctx, userID, *transactionID)
		if err != nil {
			return uuid.Nil, err
		}
		if amount <= 0 {
			amount = trans.Amount
		}
		if contributionDate == nil {
			dt := trans.CompletedAt
			contributionDate = &dt
		}
	}

	contribution, err := domain.NewGoalContribution(userID, goalID, amount, contributionDate, transactionID)
	if err != nil {
		return uuid.Nil, err
	}

	var contributionID uuid.UUID
	err = uc.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		var addErr error
		contributionID, addErr = uc.repo.AddContribution(txCtx, contribution)
		if addErr != nil {
			return addErr
		}
		return uc.repo.IncreaseCurrentAmount(txCtx, userID, goalID, amount)
	})
	if err != nil {
		return uuid.Nil, err
	}

	return contributionID, nil
}

func buildSummary(goal domain.Goal, now time.Time) domain.GoalSummary {
	progress := 0.0
	if goal.TargetAmount > 0 {
		progress = (float64(goal.CurrentAmount) / float64(goal.TargetAmount)) * 100
	}

	status := domain.GoalStatusInProgress
	if goal.CurrentAmount >= goal.TargetAmount {
		status = domain.GoalStatusAchieved
	} else if goal.TargetDate != nil {
		deadline := time.Date(goal.TargetDate.Year(), goal.TargetDate.Month(), goal.TargetDate.Day(), 23, 59, 59, 0, time.UTC)
		if now.After(deadline) {
			status = domain.GoalStatusOverdue
		}
	}

	return domain.GoalSummary{
		GoalID:          goal.GoalID,
		NameGoal:        goal.NameGoal,
		TargetAmount:    goal.TargetAmount,
		CurrentAmount:   goal.CurrentAmount,
		ProgressPercent: progress,
		Status:          status,
		TargetDate:      goal.TargetDate,
	}
}

func (uc *GoalUseCase) buildForecast(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, currentAmount int64, targetAmount int64, historyMonths int) (domain.GoalForecast, error) {
	if currentAmount >= targetAmount {
		return domain.GoalForecast{}, nil
	}

	now := time.Now().UTC()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	windowStart := currentMonthStart.AddDate(0, -historyMonths, 0)

	rows, err := uc.repo.GetMonthlyContributions(ctx, userID, goalID, windowStart, currentMonthStart)
	if err != nil {
		return domain.GoalForecast{}, err
	}

	sum := int64(0)
	for _, row := range rows {
		sum += row.TotalAmount
	}

	avg := int64(math.Round(float64(sum) / float64(historyMonths)))
	if avg <= 0 {
		return domain.GoalForecast{AverageMonthlyContribution: 0}, nil
	}

	remaining := targetAmount - currentAmount
	remainingMonths := int(math.Ceil(float64(remaining) / float64(avg)))
	if remainingMonths < 1 {
		remainingMonths = 1
	}
	estimatedDate := currentMonthStart.AddDate(0, remainingMonths, 0)

	return domain.GoalForecast{
		AverageMonthlyContribution: avg,
		EstimatedReachDate:         &estimatedDate,
		RemainingMonths:            &remainingMonths,
	}, nil
}
