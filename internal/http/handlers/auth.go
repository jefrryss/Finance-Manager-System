package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/config"
	"expenses-backend/internal/models"
	"expenses-backend/internal/services"
	"expenses-backend/internal/utils"
)

type AuthHandler struct {
	Cfg config.Config
	DB  *gorm.DB
}

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	// check exists
	var existing models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		utils.Conflict(c, "email already registered")
		return
	} else if err != nil && err != gorm.ErrRecordNotFound {
		utils.Internal(c, "db error")
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Internal(c, "hash error")
		return
	}

	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		// создаём системные категории "Другое" для income/expense
		if _, err := services.EnsureSystemOther(tx, user.ID, models.CategoryTypeExpense); err != nil {
			return err
		}
		if _, err := services.EnsureSystemOther(tx, user.ID, models.CategoryTypeIncome); err != nil {
			return err
		}
		return nil
	}); err != nil {
		utils.Internal(c, "db error")
		return
	}

	token, _, exp, err := utils.NewToken(user.ID, h.Cfg.JWTSecret, time.Duration(h.Cfg.JWTTTLHours)*time.Hour)
	if err != nil {
		utils.Internal(c, "token error")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":       gin.H{"id": user.ID, "email": user.Email},
		"token":      token,
		"expires_at": exp,
	})
}

func (h AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.Unauthorized(c, "invalid credentials")
		return
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		utils.Unauthorized(c, "invalid credentials")
		return
	}

	token, _, exp, err := utils.NewToken(user.ID, h.Cfg.JWTSecret, time.Duration(h.Cfg.JWTTTLHours)*time.Hour)
	if err != nil {
		utils.Internal(c, "token error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       gin.H{"id": user.ID, "email": user.Email},
		"token":      token,
		"expires_at": exp,
	})
}

func (h AuthHandler) Logout(c *gin.Context) {
	uidAny, _ := c.Get("user_id")
	jtiAny, _ := c.Get("token_jti")
	expAny, _ := c.Get("token_exp")

	userID, _ := uidAny.(uuid.UUID)
	jti, _ := jtiAny.(string)
	exp, _ := expAny.(time.Time)

	if userID == uuid.Nil || jti == "" {
		utils.Unauthorized(c, "invalid token")
		return
	}

	// store revoked token
	rt := models.RevokedToken{
		UserID:    userID,
		JTI:       jti,
		ExpiresAt: exp,
	}
	// ignore duplicate (already logged out)
	if err := h.DB.Create(&rt).Error; err != nil {
		// if unique constraint violation, still ok
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
