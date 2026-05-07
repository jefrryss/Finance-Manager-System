package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTransEmptyUserID     = errors.New("user ID cannot be empty")
	ErrTransEmptyAccountID  = errors.New("account ID cannot be empty")
	ErrTransEmptyName       = errors.New("transaction name cannot be empty")
	ErrTransInvalidAmount   = errors.New("amount must be strictly greater than zero")
	ErrTransNotFound        = errors.New("transaction not found")
	ErrCannotModifyImported = errors.New("cannot modify amount, date, or type of imported transactions")
)

type Transaction struct {
	TransactionID   uuid.UUID  `db:"transaction_id" json:"transaction_id"`
	UserID          uuid.UUID  `db:"user_id" json:"user_id"`
	AccountID       uuid.UUID  `db:"account_id" json:"account_id"`
	CategoryID      *uuid.UUID `db:"category_id" json:"category_id"`
	NameTransaction string     `db:"name_transaction" json:"name_transaction"`
	IsIncome        bool       `db:"is_income" json:"is_income"`
	Amount          int64      `db:"amount" json:"amount"`
	CompletedAt     time.Time  `db:"completed_at" json:"completed_at"`
	IsHidden        bool       `db:"is_hidden" json:"is_hidden"`
	IsImported      bool       `db:"is_imported" json:"is_imported"`
	Comment         *string    `db:"comment" json:"comment,omitempty"`
}

type TransactionFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	IsIncome   *bool
	StartDate  *time.Time
	EndDate    *time.Time
	IsHidden   *bool
}

func NewTransaction(
	userID uuid.UUID,
	accountID uuid.UUID,
	categoryID *uuid.UUID,
	name string,
	isIncome bool,
	amount int64,
	completedAt time.Time,
	isImported bool,
	comment *string,
) (*Transaction, error) {

	if userID == uuid.Nil {
		return nil, ErrTransEmptyUserID
	}
	if accountID == uuid.Nil {
		return nil, ErrTransEmptyAccountID
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTransEmptyName
	}

	if amount <= 0 {
		return nil, ErrTransInvalidAmount
	}

	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}

	if comment != nil {
		cleanedComment := strings.TrimSpace(*comment)
		if cleanedComment == "" {
			comment = nil
		} else {
			comment = &cleanedComment
		}
	}

	return &Transaction{
		TransactionID:   uuid.Nil,
		UserID:          userID,
		AccountID:       accountID,
		CategoryID:      categoryID,
		NameTransaction: name,
		IsIncome:        isIncome,
		Amount:          amount,
		CompletedAt:     completedAt,
		IsHidden:        false,
		IsImported:      isImported,
		Comment:         comment,
	}, nil
}
