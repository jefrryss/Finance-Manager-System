package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountType string

const (
	AccountTypeManual   AccountType = "manual"
	AccountTypeImported AccountType = "imported"
)

type Account struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID   `json:"user_id" gorm:"type:uuid;index;not null"`
	Name           string      `json:"name" gorm:"not null"`
	Type           AccountType `json:"type" gorm:"type:varchar(16);not null;default:'manual'"`
	InitialBalance int64       `json:"initial_balance" gorm:"not null;default:0"`
	Balance        int64       `json:"balance" gorm:"not null;default:0"`

	// служебные поля для внешней интеграции (можно не использовать в курсовом, но есть в модели)
	ExternalID   *string    `json:"external_id,omitempty" gorm:"type:varchar(128);index"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
