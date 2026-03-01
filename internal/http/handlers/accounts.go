package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/http/middleware"
	"expenses-backend/internal/models"
	"expenses-backend/internal/utils"
)

type AccountsHandler struct {
	DB *gorm.DB
}

type createAccountReq struct {
	Name           string              `json:"name" binding:"required"`
	Type           *models.AccountType `json:"type,omitempty"` // manual/imported
	InitialBalance int64               `json:"initial_balance"`
	ExternalID     *string             `json:"external_id,omitempty"`
	LastSyncedAt   *time.Time          `json:"last_synced_at,omitempty"`
}

func (h AccountsHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		utils.Unauthorized(c, "unauthorized")
		return
	}

	var req createAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	typeVal := models.AccountTypeManual
	if req.Type != nil {
		typeVal = *req.Type
	}
	if typeVal != models.AccountTypeManual && typeVal != models.AccountTypeImported {
		utils.BadRequest(c, "invalid account type", "type must be 'manual' or 'imported'")
		return
	}

	acc := models.Account{
		UserID:         userID,
		Name:           req.Name,
		Type:           typeVal,
		InitialBalance: req.InitialBalance,
		Balance:        req.InitialBalance,
		ExternalID:     req.ExternalID,
		LastSyncedAt:   req.LastSyncedAt,
	}

	if err := h.DB.Create(&acc).Error; err != nil {
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"account": acc})
}

func (h AccountsHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	var accounts []models.Account
	if err := h.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&accounts).Error; err != nil {
		utils.Internal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func (h AccountsHandler) Get(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var acc models.Account
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&acc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "account not found")
			return
		}
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": acc})
}

type patchAccountReq struct {
	Name *string `json:"name,omitempty"`
}

func (h AccountsHandler) Patch(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var req patchAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}
	if req.Name == nil || *req.Name == "" {
		utils.BadRequest(c, "nothing to update", nil)
		return
	}

	res := h.DB.Model(&models.Account{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]interface{}{
		"name": *req.Name,
	})
	if res.Error != nil {
		utils.Internal(c, "db error")
		return
	}
	if res.RowsAffected == 0 {
		utils.NotFound(c, "account not found")
		return
	}

	var acc models.Account
	_ = h.DB.Where("id = ?", id).First(&acc).Error
	c.JSON(http.StatusOK, gin.H{"account": acc})
}

func (h AccountsHandler) Delete(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var acc models.Account
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&acc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "account not found")
			return
		}
		utils.Internal(c, "db error")
		return
	}

	if acc.Type == models.AccountTypeImported {
		utils.Conflict(c, "imported accounts cannot be deleted")
		return
	}

	// delete account + transactions (cascade manually)
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id = ? AND user_id = ?", acc.ID, userID).Delete(&models.Transaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND user_id = ?", acc.ID, userID).Delete(&models.Account{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
