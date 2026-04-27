package domain

import (
	"time"

	"github.com/google/uuid"
)

type SummaryReport struct {
	TotalIncome  int64 `db:"total_income" json:"total_income"`
	TotalExpense int64 `db:"total_expense" json:"total_expense"`
}

type CategoryReport struct {
	CategoryID   *uuid.UUID `db:"category_id" json:"category_id"`
	CategoryName string     `db:"category_name" json:"category_name"`
	IconURL      *string    `db:"icon_url" json:"icon_url,omitempty"`
	TotalAmount  int64      `db:"total_amount" json:"total_amount"`
	SharePercent float64    `db:"share_percent" json:"share_percent"`
}

type DailyReport struct {
	Date        time.Time `db:"date" json:"date"`
	TotalAmount int64     `db:"total_amount" json:"total_amount"`
}

type MonthlyReport struct {
	Month       time.Time `db:"month" json:"month"`
	TotalAmount int64     `db:"total_amount" json:"total_amount"`
}

type CategoryCompareReport struct {
	CategoryID         *uuid.UUID `json:"category_id"`
	CategoryName       string     `json:"category_name"`
	IconURL            *string    `json:"icon_url,omitempty"`
	FirstPeriodAmount  int64      `json:"first_period_amount"`
	SecondPeriodAmount int64      `json:"second_period_amount"`
	DeltaAmount        int64      `json:"delta_amount"`
	DeltaPercent       *float64   `json:"delta_percent,omitempty"`
}
