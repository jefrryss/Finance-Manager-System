package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrCatEmptyUserID = errors.New("user ID cannot be empty (nil UUID)")
	ErrCatEmptyName   = errors.New("category name cannot be empty")
	ErrCatNameLong    = errors.New("category name cannot be longer than 255 characters")
)

type Category struct {
	CategoryID   uuid.UUID `db:"category_id" json:"category_id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	NameCategory string    `db:"name_category" json:"name_category"`
	IsIncome     bool      `db:"is_income" json:"is_income"`
	IsCustom     bool      `db:"is_custom" json:"is_custom"`
	IconURL      *string   `db:"icon_url" json:"icon_url,omitempty"`
}

func NewCategory(
	userID uuid.UUID,
	name string,
	isIncome bool,
	isCustom bool,
	iconURL *string,
) (*Category, error) {

	if userID == uuid.Nil {
		return nil, ErrCatEmptyUserID
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrCatEmptyName
	}
	if len([]rune(name)) > 255 {
		return nil, ErrCatNameLong
	}

	var finalIconURL *string
	if iconURL != nil {
		cleaned := strings.TrimSpace(*iconURL)
		if cleaned != "" {
			finalIconURL = &cleaned
		}
	}

	return &Category{
		CategoryID:   uuid.Nil,
		UserID:       userID,
		NameCategory: name,
		IsIncome:     isIncome,
		IsCustom:     isCustom,
		IconURL:      finalIconURL,
	}, nil
}
