package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"expenses-backend/internal/models"
)

var ErrNoSuchAccount = errors.New("account not found")

func GetAccountForUser(db *gorm.DB, userID, accountID uuid.UUID) (models.Account, error) {
	var acc models.Account
	err := db.Where("id = ? AND user_id = ?", accountID, userID).First(&acc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Account{}, ErrNoSuchAccount
		}
		return models.Account{}, err
	}
	return acc, nil
}

func ApplyDeltaToAccountBalance(tx *gorm.DB, accountID uuid.UUID, delta int64) error {
	// optimistic: UPDATE accounts SET balance = balance + ? WHERE id = ?
	res := tx.Model(&models.Account{}).
		Where("id = ?", accountID).
		UpdateColumn("balance", gorm.Expr("balance + ?", delta))
	return res.Error
}

func DeltaForTransaction(t models.Transaction) int64 {
	// income increases balance, expense decreases
	if t.Type == models.TxTypeIncome {
		return t.Amount
	}
	return -t.Amount
}
