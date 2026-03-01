package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TxType string

const (
	TxTypeIncome  TxType = "income"
	TxTypeExpense TxType = "expense"
)

type Transaction struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;index;not null"`
	AccountID  uuid.UUID `json:"account_id" gorm:"type:uuid;index;not null"`
	CategoryID uuid.UUID `json:"category_id" gorm:"type:uuid;index;not null"`

	Amount     int64     `json:"amount" gorm:"not null"` // int64 (копейки/центы)
	Type       TxType    `json:"type" gorm:"type:varchar(16);not null"`
	OccurredAt time.Time `json:"occurred_at" gorm:"index;not null"`
	Comment    *string   `json:"comment,omitempty" gorm:"type:text"`

	IsImported bool `json:"is_imported" gorm:"not null;default:false"`
	IsHidden   bool `json:"is_hidden" gorm:"not null;default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
