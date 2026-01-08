package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

type Category struct {
	ID        uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID   `json:"user_id,omitempty" gorm:"type:uuid;index"` // nil => общая категория
	Name      string       `json:"name" gorm:"not null"`
	Type      CategoryType `json:"type" gorm:"type:varchar(16);not null"`
	IsSystem  bool         `json:"is_system" gorm:"not null;default:false"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
