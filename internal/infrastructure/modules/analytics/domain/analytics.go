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
	CategoryID  *uuid.UUID `db:"category_id" json:"category_id"`
	TotalAmount int64      `db:"total_amount" json:"total_amount"`
}

type DailyReport struct {
	Date        time.Time `db:"date" json:"date"`
	TotalAmount int64     `db:"total_amount" json:"total_amount"`
}
