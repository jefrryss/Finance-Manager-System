package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyUserID       = errors.New("user ID cannot be empty (nil UUID)")
	ErrEmptyAccountName  = errors.New("account name cannot be empty")
	ErrAccountNameLong   = errors.New("account name cannot be longer than 50 characters")
	ErrInvalidCurrency   = errors.New("currency must be a valid 3-letter ISO code")
	ErrEmptyAccountType  = errors.New("account type cannot be empty")
	ErrInvalidColorHex   = errors.New("color must be a valid hex code (e.g., #FFFFFF)")
	ErrMissingExternalID = errors.New("external account ID is required for imported accounts")
)

type Account struct {
	AccountID         uuid.UUID  `db:"account_id" json:"account_id"`
	UserID            uuid.UUID  `db:"user_id" json:"user_id"`
	Balance           int64      `db:"balance" json:"balance"`
	IsImported        bool       `db:"is_imported" json:"is_imported"`
	ExternalAccountID *string    `db:"external_account_id" json:"external_account_id,omitempty"`
	AccountType       string     `db:"account_type" json:"account_type"`
	ColorHex          string     `db:"color_hex" json:"color_hex"`
	IsArchived        bool       `db:"is_archived" json:"is_archived"`
	NameAccount       string     `db:"name_account" json:"name_account"`
	Currency          string     `db:"currency" json:"currency"`
	LastSyncedAt      *time.Time `db:"last_synced_at" json:"last_synced_at,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}

func NewAccount(
	userID uuid.UUID,
	name string,
	currency string,
	accountType string,
	colorHex string,
	isImported bool,
	externalAccountID *string,
	initialBalance int64,
) (*Account, error) {

	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyAccountName
	}
	if len([]rune(name)) > 50 {
		return nil, ErrAccountNameLong
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}

	accountType = strings.ToUpper(strings.TrimSpace(accountType))
	if accountType == "" {
		return nil, ErrEmptyAccountType
	}

	colorHex = strings.ToUpper(strings.TrimSpace(colorHex))
	if colorHex != "" {
		if len(colorHex) != 7 || !strings.HasPrefix(colorHex, "#") {
			return nil, ErrInvalidColorHex
		}
	} else {
		colorHex = "#000000"
	}

	if isImported {
		if externalAccountID == nil || strings.TrimSpace(*externalAccountID) == "" {
			return nil, ErrMissingExternalID
		}
	} else {
		externalAccountID = nil
	}

	return &Account{
		AccountID:         uuid.Nil,
		UserID:            userID,
		Balance:           initialBalance,
		IsImported:        isImported,
		ExternalAccountID: externalAccountID,
		AccountType:       accountType,
		ColorHex:          colorHex,
		IsArchived:        false,
		NameAccount:       name,
		Currency:          currency,
		LastSyncedAt:      nil,
		CreatedAt:         time.Now().UTC(),
	}, nil
}
