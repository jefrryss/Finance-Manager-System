package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGoalEmptyUserID               = errors.New("user ID cannot be empty (nil UUID)")
	ErrGoalEmptyName                 = errors.New("goal name cannot be empty")
	ErrGoalNameTooLong               = errors.New("goal name cannot be longer than 255 characters")
	ErrGoalInvalidTargetAmount       = errors.New("target amount must be strictly greater than zero")
	ErrGoalInvalidContributionAmount = errors.New("contribution amount must be strictly greater than zero")
	ErrGoalNotFound                  = errors.New("goal not found")
)

type GoalStatus string

const (
	GoalStatusInProgress GoalStatus = "in_progress"
	GoalStatusAchieved   GoalStatus = "achieved"
	GoalStatusOverdue    GoalStatus = "overdue"
)

type Goal struct {
	GoalID        uuid.UUID  `db:"goal_id" json:"goal_id"`
	UserID        uuid.UUID  `db:"user_id" json:"user_id"`
	NameGoal      string     `db:"name_goal" json:"name_goal"`
	TargetAmount  int64      `db:"target_amount" json:"target_amount"`
	CurrentAmount int64      `db:"current_amount" json:"current_amount"`
	TargetDate    *time.Time `db:"target_date" json:"target_date,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

type GoalContribution struct {
	ContributionID   uuid.UUID  `db:"contribution_id" json:"contribution_id"`
	GoalID           uuid.UUID  `db:"goal_id" json:"goal_id"`
	UserID           uuid.UUID  `db:"user_id" json:"user_id"`
	Amount           int64      `db:"amount" json:"amount"`
	ContributionDate time.Time  `db:"contribution_date" json:"contribution_date"`
	TransactionID    *uuid.UUID `db:"transaction_id" json:"transaction_id,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}

type GoalSummary struct {
	GoalID          uuid.UUID  `json:"goal_id"`
	NameGoal        string     `json:"name_goal"`
	TargetAmount    int64      `json:"target_amount"`
	CurrentAmount   int64      `json:"current_amount"`
	ProgressPercent float64    `json:"progress_percent"`
	Status          GoalStatus `json:"status"`
	TargetDate      *time.Time `json:"target_date,omitempty"`
}

type GoalForecast struct {
	AverageMonthlyContribution int64      `json:"average_monthly_contribution"`
	EstimatedReachDate         *time.Time `json:"estimated_reach_date,omitempty"`
	RemainingMonths            *int       `json:"remaining_months,omitempty"`
}

type GoalDetails struct {
	Summary       GoalSummary        `json:"summary"`
	Contributions []GoalContribution `json:"contributions"`
	Forecast      GoalForecast       `json:"forecast"`
}

type MonthlyContribution struct {
	Month       time.Time `db:"month"`
	TotalAmount int64     `db:"total_amount"`
}

func NewGoal(userID uuid.UUID, name string, targetAmount int64, targetDate *time.Time) (*Goal, error) {
	if userID == uuid.Nil {
		return nil, ErrGoalEmptyUserID
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrGoalEmptyName
	}
	if len([]rune(name)) > 255 {
		return nil, ErrGoalNameTooLong
	}
	if targetAmount <= 0 {
		return nil, ErrGoalInvalidTargetAmount
	}

	var cleanedDate *time.Time
	if targetDate != nil {
		t := targetDate.UTC()
		cleanedDate = &t
	}

	now := time.Now().UTC()

	return &Goal{
		GoalID:        uuid.Nil,
		UserID:        userID,
		NameGoal:      name,
		TargetAmount:  targetAmount,
		CurrentAmount: 0,
		TargetDate:    cleanedDate,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func NewGoalContribution(userID uuid.UUID, goalID uuid.UUID, amount int64, contributionDate *time.Time, transactionID *uuid.UUID) (*GoalContribution, error) {
	if userID == uuid.Nil {
		return nil, ErrGoalEmptyUserID
	}
	if goalID == uuid.Nil {
		return nil, ErrGoalNotFound
	}
	if amount <= 0 {
		return nil, ErrGoalInvalidContributionAmount
	}

	finalDate := time.Now().UTC()
	if contributionDate != nil && !contributionDate.IsZero() {
		finalDate = contributionDate.UTC()
	}

	return &GoalContribution{
		ContributionID:   uuid.Nil,
		GoalID:           goalID,
		UserID:           userID,
		Amount:           amount,
		ContributionDate: finalDate,
		TransactionID:    transactionID,
		CreatedAt:        time.Now().UTC(),
	}, nil
}
