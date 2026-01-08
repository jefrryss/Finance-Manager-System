package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/config"
	"expenses-backend/internal/models"
	"expenses-backend/internal/utils"
)

const CtxUserIDKey = "user_id"

func AuthRequired(cfg config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			utils.Unauthorized(c, "missing bearer token")
			return
		}

		tok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		userID, jti, exp, err := utils.ParseToken(tok, cfg.JWTSecret)
		if err != nil {
			utils.Unauthorized(c, "invalid token")
			return
		}

		// проверяем, не отозван ли токен (logout)
		var rt models.RevokedToken
		err = db.Where("jti = ?", jti).First(&rt).Error
		if err == nil {
			utils.Unauthorized(c, "token revoked")
			return
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			utils.Internal(c, "db error")
			return
		}

		// если exp пустой (не должен), ставим короткий TTL, чтобы не засорять revoked
		if exp.IsZero() {
			exp = time.Now().UTC().Add(1 * time.Hour)
		}

		c.Set(CtxUserIDKey, userID)
		c.Set("token_jti", jti)
		c.Set("token_exp", exp)
		c.Next()
	}
}

func MustGetUserID(c *gin.Context) uuid.UUID {
	v, ok := c.Get(CtxUserIDKey)
	if !ok {
		return uuid.Nil
	}
	uid, _ := v.(uuid.UUID)
	return uid
}
