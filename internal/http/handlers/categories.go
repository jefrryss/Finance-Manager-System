package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/http/middleware"
	"expenses-backend/internal/models"
	"expenses-backend/internal/services"
	"expenses-backend/internal/utils"
)

type CategoriesHandler struct {
	DB *gorm.DB
}

func (h CategoriesHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	var cats []models.Category
	if err := h.DB.Where("(user_id IS NULL) OR (user_id = ?)", userID).Order("is_system desc, created_at desc").Find(&cats).Error; err != nil {
		utils.Internal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

type createCategoryReq struct {
	Name string              `json:"name" binding:"required"`
	Type models.CategoryType `json:"type" binding:"required"`
}

func (h CategoriesHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req createCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}
	if req.Type != models.CategoryTypeIncome && req.Type != models.CategoryTypeExpense {
		utils.BadRequest(c, "invalid category type", "type must be 'income' or 'expense'")
		return
	}

	uid := userID
	cat := models.Category{
		UserID:   &uid,
		Name:     req.Name,
		Type:     req.Type,
		IsSystem: false,
	}
	if err := h.DB.Create(&cat).Error; err != nil {
		utils.Internal(c, "db error")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": cat})
}

type patchCategoryReq struct {
	Name *string `json:"name,omitempty"`
}

func (h CategoriesHandler) Patch(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var req patchCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}
	if req.Name == nil || *req.Name == "" {
		utils.BadRequest(c, "nothing to update", nil)
		return
	}

	// редактировать можно только пользовательские категории (user_id = current user, is_system=false)
	res := h.DB.Model(&models.Category{}).
		Where("id = ? AND user_id = ? AND is_system = false", id, userID).
		Update("name", *req.Name)

	if res.Error != nil {
		utils.Internal(c, "db error")
		return
	}
	if res.RowsAffected == 0 {
		utils.NotFound(c, "category not found or not editable")
		return
	}

	var cat models.Category
	_ = h.DB.Where("id = ?", id).First(&cat).Error
	c.JSON(http.StatusOK, gin.H{"category": cat})
}

func (h CategoriesHandler) Delete(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var cat models.Category
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&cat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "category not found")
			return
		}
		utils.Internal(c, "db error")
		return
	}

	if cat.IsSystem {
		utils.Conflict(c, "system category cannot be deleted")
		return
	}

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		other, err := services.EnsureSystemOther(tx, userID, cat.Type)
		if err != nil {
			return err
		}
		// перенос транзакций в "Другое"
		if err := tx.Model(&models.Transaction{}).
			Where("user_id = ? AND category_id = ?", userID, cat.ID).
			Update("category_id", other.ID).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND user_id = ?", cat.ID, userID).Delete(&models.Category{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
