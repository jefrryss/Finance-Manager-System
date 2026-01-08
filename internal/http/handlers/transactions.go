package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/http/middleware"
	"expenses-backend/internal/models"
	"expenses-backend/internal/services"
	"expenses-backend/internal/utils"
)

type TransactionsHandler struct {
	DB *gorm.DB
}

type createTxReq struct {
	AccountID  uuid.UUID     `json:"account_id" binding:"required"`
	CategoryID *uuid.UUID    `json:"category_id,omitempty"`
	Amount     int64         `json:"amount" binding:"required"`
	Type       models.TxType `json:"type" binding:"required"`
	OccurredAt time.Time     `json:"occurred_at" binding:"required"`
	Comment    *string       `json:"comment,omitempty"`
	IsImported *bool         `json:"is_imported,omitempty"`
	IsHidden   *bool         `json:"is_hidden,omitempty"`
}

func (h TransactionsHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req createTxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}
	if req.Amount < 0 {
		utils.BadRequest(c, "amount must be non-negative", nil)
		return
	}
	if req.Type != models.TxTypeIncome && req.Type != models.TxTypeExpense {
		utils.BadRequest(c, "invalid tx type", "type must be 'income' or 'expense'")
		return
	}

	// account ownership
	acc, err := services.GetAccountForUser(h.DB, userID, req.AccountID)
	if err != nil {
		utils.NotFound(c, "account not found")
		return
	}

	// category: if omitted => "Другое" for this type
	catID := uuid.Nil
	if req.CategoryID != nil {
		catID = *req.CategoryID
		cat, err := services.ValidateCategoryOwnership(h.DB, userID, catID)
		if err != nil {
			utils.NotFound(c, "category not found")
			return
		}
		// optional: check type match (income vs expense)
		if string(cat.Type) != string(req.Type) {
			// allow mismatch? Usually category type should match tx type
			utils.BadRequest(c, "category type mismatch", nil)
			return
		}
	} else {
		ctype := models.CategoryTypeExpense
		if req.Type == models.TxTypeIncome {
			ctype = models.CategoryTypeIncome
		}
		other, err := services.EnsureSystemOther(h.DB, userID, ctype)
		if err != nil {
			utils.Internal(c, "db error")
			return
		}
		catID = other.ID
	}

	isImported := false
	if req.IsImported != nil {
		isImported = *req.IsImported
	}
	isHidden := false
	if req.IsHidden != nil {
		isHidden = *req.IsHidden
	}

	txModel := models.Transaction{
		UserID:     userID,
		AccountID:  acc.ID,
		CategoryID: catID,
		Amount:     req.Amount,
		Type:       req.Type,
		OccurredAt: req.OccurredAt.UTC(),
		Comment:    req.Comment,
		IsImported: isImported,
		IsHidden:   isHidden,
	}

	if err := h.DB.Transaction(func(dbtx *gorm.DB) error {
		if err := dbtx.Create(&txModel).Error; err != nil {
			return err
		}
		delta := services.DeltaForTransaction(txModel)
		return services.ApplyDeltaToAccountBalance(dbtx, acc.ID, delta)
	}); err != nil {
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"transaction": txModel})
}

func (h TransactionsHandler) Get(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var txModel models.Transaction
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&txModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "transaction not found")
			return
		}
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txModel})
}

func (h TransactionsHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	q := h.DB.Model(&models.Transaction{}).Where("user_id = ?", userID)

	if v := c.Query("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			utils.BadRequest(c, "invalid account_id", nil)
			return
		}
		q = q.Where("account_id = ?", id)
	}
	if v := c.Query("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			utils.BadRequest(c, "invalid category_id", nil)
			return
		}
		q = q.Where("category_id = ?", id)
	}
	if v := c.Query("type"); v != "" {
		if v != string(models.TxTypeIncome) && v != string(models.TxTypeExpense) {
			utils.BadRequest(c, "invalid type", nil)
			return
		}
		q = q.Where("type = ?", v)
	}
	if v := c.Query("is_hidden"); v != "" {
		if v == "true" || v == "1" {
			q = q.Where("is_hidden = true")
		} else if v == "false" || v == "0" {
			q = q.Where("is_hidden = false")
		} else {
			utils.BadRequest(c, "invalid is_hidden", nil)
			return
		}
	}
	if v := c.Query("start"); v != "" {
		tm, err := time.Parse(time.RFC3339, v)
		if err != nil {
			utils.BadRequest(c, "invalid start (use RFC3339)", nil)
			return
		}
		q = q.Where("occurred_at >= ?", tm.UTC())
	}
	if v := c.Query("end"); v != "" {
		tm, err := time.Parse(time.RFC3339, v)
		if err != nil {
			utils.BadRequest(c, "invalid end (use RFC3339)", nil)
			return
		}
		q = q.Where("occurred_at <= ?", tm.UTC())
	}

	var out []models.Transaction
	if err := q.Order("occurred_at desc").Find(&out).Error; err != nil {
		utils.Internal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": out})
}

type patchTxReq struct {
	Amount     *int64         `json:"amount,omitempty"`
	Type       *models.TxType `json:"type,omitempty"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	CategoryID *uuid.UUID     `json:"category_id,omitempty"`
	Comment    *string        `json:"comment,omitempty"`
	IsHidden   *bool          `json:"is_hidden,omitempty"`
}

func (h TransactionsHandler) Patch(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var req patchTxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	var txModel models.Transaction
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&txModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "transaction not found")
			return
		}
		utils.Internal(c, "db error")
		return
	}

	// Imported: can update only category/comment/is_hidden
	if txModel.IsImported {
		updates := map[string]interface{}{}
		if req.CategoryID != nil {
			cat, err := services.ValidateCategoryOwnership(h.DB, userID, *req.CategoryID)
			if err != nil {
				utils.NotFound(c, "category not found")
				return
			}
			if string(cat.Type) != string(txModel.Type) {
				utils.BadRequest(c, "category type mismatch", nil)
				return
			}
			updates["category_id"] = *req.CategoryID
		}
		if req.Comment != nil {
			updates["comment"] = req.Comment
		}
		if req.IsHidden != nil {
			updates["is_hidden"] = *req.IsHidden
		}
		if len(updates) == 0 {
			utils.BadRequest(c, "nothing to update", nil)
			return
		}
		if err := h.DB.Model(&models.Transaction{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates).Error; err != nil {
			utils.Internal(c, "db error")
			return
		}
		_ = h.DB.Where("id = ?", id).First(&txModel).Error
		c.JSON(http.StatusOK, gin.H{"transaction": txModel})
		return
	}

	// Manual: can update amount/type/occurred_at/category/comment/is_hidden
	newModel := txModel
	if req.Amount != nil {
		if *req.Amount < 0 {
			utils.BadRequest(c, "amount must be non-negative", nil)
			return
		}
		newModel.Amount = *req.Amount
	}
	if req.Type != nil {
		if *req.Type != models.TxTypeIncome && *req.Type != models.TxTypeExpense {
			utils.BadRequest(c, "invalid type", nil)
			return
		}
		newModel.Type = *req.Type
	}
	if req.OccurredAt != nil {
		newModel.OccurredAt = req.OccurredAt.UTC()
	}
	if req.CategoryID != nil {
		cat, err := services.ValidateCategoryOwnership(h.DB, userID, *req.CategoryID)
		if err != nil {
			utils.NotFound(c, "category not found")
			return
		}
		// check category type matches newModel.Type
		if string(cat.Type) != string(newModel.Type) {
			utils.BadRequest(c, "category type mismatch", nil)
			return
		}
		newModel.CategoryID = *req.CategoryID
	}
	if req.Comment != nil {
		newModel.Comment = req.Comment
	}
	if req.IsHidden != nil {
		newModel.IsHidden = *req.IsHidden
	}

	oldDelta := services.DeltaForTransaction(txModel)
	newDelta := services.DeltaForTransaction(newModel)
	deltaDiff := newDelta - oldDelta

	if err := h.DB.Transaction(func(dbtx *gorm.DB) error {
		if err := dbtx.Model(&models.Transaction{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]interface{}{
			"amount":      newModel.Amount,
			"type":        newModel.Type,
			"occurred_at": newModel.OccurredAt,
			"category_id": newModel.CategoryID,
			"comment":     newModel.Comment,
			"is_hidden":   newModel.IsHidden,
		}).Error; err != nil {
			return err
		}
		if deltaDiff != 0 {
			return services.ApplyDeltaToAccountBalance(dbtx, newModel.AccountID, deltaDiff)
		}
		return nil
	}); err != nil {
		utils.Internal(c, "db error")
		return
	}

	_ = h.DB.Where("id = ?", id).First(&txModel).Error
	c.JSON(http.StatusOK, gin.H{"transaction": txModel})
}

func (h TransactionsHandler) Delete(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid id", nil)
		return
	}

	var txModel models.Transaction
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&txModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "transaction not found")
			return
		}
		utils.Internal(c, "db error")
		return
	}

	if txModel.IsImported {
		utils.Conflict(c, "imported transactions cannot be deleted")
		return
	}

	if err := h.DB.Transaction(func(dbtx *gorm.DB) error {
		if err := dbtx.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Transaction{}).Error; err != nil {
			return err
		}
		delta := -services.DeltaForTransaction(txModel)
		return services.ApplyDeltaToAccountBalance(dbtx, txModel.AccountID, delta)
	}); err != nil {
		utils.Internal(c, "db error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
