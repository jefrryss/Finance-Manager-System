package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/models"
)

var ErrNoSuchCategory = errors.New("category not found")

// EnsureSystemOther creates (if missing) "Другое" category for user + type.
func EnsureSystemOther(db *gorm.DB, userID uuid.UUID, ctype models.CategoryType) (models.Category, error) {
	var cat models.Category
	err := db.Where("user_id = ? AND type = ? AND is_system = true AND name = ?", userID, ctype, "Другое").
		First(&cat).Error
	if err == nil {
		return cat, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return models.Category{}, err
	}

	uid := userID
	cat = models.Category{
		UserID:   &uid,
		Name:     "Другое",
		Type:     ctype,
		IsSystem: true,
	}
	if err := db.Create(&cat).Error; err != nil {
		return models.Category{}, err
	}
	return cat, nil
}

// ValidateCategoryOwnership checks that category exists and is visible for the user (global or owned).
func ValidateCategoryOwnership(db *gorm.DB, userID uuid.UUID, categoryID uuid.UUID) (models.Category, error) {
	var cat models.Category
	err := db.Where("id = ? AND (user_id IS NULL OR user_id = ?)", categoryID, userID).First(&cat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Category{}, ErrNoSuchCategory
		}
		return models.Category{}, err
	}
	return cat, nil
}
