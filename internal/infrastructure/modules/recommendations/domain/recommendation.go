package domain

import (
	"time"

	"github.com/google/uuid"
)

type BudgetRecommendation struct {
	CategoryID        *uuid.UUID `json:"category_id,omitempty"`
	CategoryName      string     `json:"category_name"`
	IconURL           *string    `json:"icon_url,omitempty"`
	AverageShare      float64    `json:"average_share"`
	RecommendedLimit  int64      `json:"recommended_limit"`
	LastMonthExpense  int64      `json:"last_month_expense"`
	IsOverBudget      bool       `json:"is_over_budget"`
	OverBudgetAmount  int64      `json:"over_budget_amount"`
	OverBudgetPercent float64    `json:"over_budget_percent"`
}

type MonthlyCategoryExpense struct {
	Month        time.Time  `db:"month"`
	CategoryID   *uuid.UUID `db:"category_id"`
	CategoryName string     `db:"category_name"`
	IconURL      *string    `db:"icon_url"`
	Amount       int64      `db:"amount"`
}
