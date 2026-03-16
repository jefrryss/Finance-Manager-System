package domain

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	UserID          uuid.UUID  `db:"user_id" json:"user_id"`
	TransactionID   int        `db:"transaction_id" json:"transaction_id"`
	AccountID       int        `db:"account_id" json:"account_id"`
	CategoryID      *uuid.UUID `db:"category_id" json:"category_id"`
	NameTransaction string     `db:"name_transaction" json:"name_transaction"`
	IsIncome        bool       `db:"is_income" json:"is_income"`
	Amount          int64      `db:"amount" json:"amount"`
	CompletedAt     time.Time  `db:"completed_at" json:"completed_at"`
	IsHidden        bool       `db:"is_hidden" json:"is_hidden"`
	IsImported      bool       `db:"is_imported" json:"is_imported"`
	Comment         *string    `db:"comment" json:"comment,omitempty"`
}
