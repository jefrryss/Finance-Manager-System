package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/goals/domain"
)

type GoalRepo struct {
	db *sqlx.DB
}

func NewGoalRepo(db *sqlx.DB) *GoalRepo {
	return &GoalRepo{db: db}
}

func (r *GoalRepo) AddGoal(ctx context.Context, goal *domain.Goal) (uuid.UUID, error) {
	q := database.GetQueryer(ctx, r.db)
	query := `
		INSERT INTO Goals (user_id, name_goal, target_amount, current_amount, target_date, created_at, updated_at)
		VALUES (:user_id, :name_goal, :target_amount, :current_amount, :target_date, :created_at, :updated_at)
		RETURNING goal_id
	`
	queryStr, args, err := sqlx.Named(query, goal)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to process named query: %w", err)
	}
	queryStr = q.Rebind(queryStr)

	var goalID uuid.UUID
	if err := q.QueryRowContext(ctx, queryStr, args...).Scan(&goalID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add goal: %w", err)
	}
	return goalID, nil
}

func (r *GoalRepo) GetGoalsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	q := database.GetQueryer(ctx, r.db)
	goals := make([]domain.Goal, 0)
	query := `SELECT * FROM Goals WHERE user_id = $1 ORDER BY created_at DESC`
	if err := q.SelectContext(ctx, &goals, query, userID); err != nil {
		return nil, fmt.Errorf("failed to get goals: %w", err)
	}
	return goals, nil
}

func (r *GoalRepo) GetGoalByID(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) (*domain.Goal, error) {
	q := database.GetQueryer(ctx, r.db)
	var goal domain.Goal
	query := `SELECT * FROM Goals WHERE user_id = $1 AND goal_id = $2`
	if err := q.GetContext(ctx, &goal, query, userID, goalID); err != nil {
		return nil, domain.ErrGoalNotFound
	}
	return &goal, nil
}

func (r *GoalRepo) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	q := database.GetQueryer(ctx, r.db)
	query := `
		UPDATE Goals
		SET name_goal = :name_goal, target_amount = :target_amount, target_date = :target_date, updated_at = :updated_at
		WHERE goal_id = :goal_id AND user_id = :user_id
	`
	res, err := q.NamedExecContext(ctx, query, goal)
	if err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrGoalNotFound
	}
	return nil
}

func (r *GoalRepo) DeleteGoal(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) error {
	q := database.GetQueryer(ctx, r.db)
	query := `DELETE FROM Goals WHERE user_id = $1 AND goal_id = $2`
	res, err := q.ExecContext(ctx, query, userID, goalID)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrGoalNotFound
	}
	return nil
}

func (r *GoalRepo) AddContribution(ctx context.Context, contribution *domain.GoalContribution) (uuid.UUID, error) {
	q := database.GetQueryer(ctx, r.db)
	query := `
		INSERT INTO GoalContributions (goal_id, user_id, amount, contribution_date, transaction_id, created_at)
		VALUES (:goal_id, :user_id, :amount, :contribution_date, :transaction_id, :created_at)
		RETURNING contribution_id
	`
	queryStr, args, err := sqlx.Named(query, contribution)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to process contribution query: %w", err)
	}
	queryStr = q.Rebind(queryStr)

	var contributionID uuid.UUID
	if err := q.QueryRowContext(ctx, queryStr, args...).Scan(&contributionID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add contribution: %w", err)
	}
	return contributionID, nil
}

func (r *GoalRepo) IncreaseCurrentAmount(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, amount int64) error {
	q := database.GetQueryer(ctx, r.db)
	query := `
		UPDATE Goals
		SET current_amount = current_amount + $1, updated_at = $2
		WHERE user_id = $3 AND goal_id = $4
	`
	res, err := q.ExecContext(ctx, query, amount, time.Now().UTC(), userID, goalID)
	if err != nil {
		return fmt.Errorf("failed to increase current amount: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrGoalNotFound
	}
	return nil
}

func (r *GoalRepo) GetGoalContributions(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) ([]domain.GoalContribution, error) {
	q := database.GetQueryer(ctx, r.db)
	contributions := make([]domain.GoalContribution, 0)
	query := `
		SELECT * FROM GoalContributions
		WHERE user_id = $1 AND goal_id = $2
		ORDER BY contribution_date DESC, created_at DESC
	`
	if err := q.SelectContext(ctx, &contributions, query, userID, goalID); err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}
	return contributions, nil
}

func (r *GoalRepo) GetMonthlyContributions(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, start time.Time, end time.Time) ([]domain.MonthlyContribution, error) {
	q := database.GetQueryer(ctx, r.db)
	rows := make([]domain.MonthlyContribution, 0)
	query := `
		SELECT
			date_trunc('month', contribution_date)::date AS month,
			COALESCE(SUM(amount), 0) AS total_amount
		FROM GoalContributions
		WHERE user_id = $1 AND goal_id = $2 AND contribution_date >= $3 AND contribution_date < $4
		GROUP BY date_trunc('month', contribution_date)::date
		ORDER BY month ASC
	`
	if err := q.SelectContext(ctx, &rows, query, userID, goalID, start, end); err != nil {
		return nil, fmt.Errorf("failed to get monthly contributions: %w", err)
	}
	return rows, nil
}
