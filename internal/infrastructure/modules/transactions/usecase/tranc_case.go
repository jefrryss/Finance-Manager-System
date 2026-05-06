package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

var (
	ErrCannotModifyImported = errors.New("cannot modify or delete imported transactions")
)

type TransactionRepository interface {
	GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*domain.Transaction, error)
	AddTransaction(ctx context.Context, trans *domain.Transaction) error
	DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error
	GetAllTransactions(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error)
	ShowTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error
	HideTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error
	GetTransactionsByIDs(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID) ([]domain.Transaction, error)
}

type AccountBalanceUpdater interface {
	UpdateBalance(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, amountDelta int64) error
}

type TransactionUseCase struct {
	transRepo   TransactionRepository
	accountRepo AccountBalanceUpdater
	txManager   database.TxManager
}

func NewTransactionUseCase(tr TransactionRepository, ar AccountBalanceUpdater, tm database.TxManager) *TransactionUseCase {
	return &TransactionUseCase{
		transRepo:   tr,
		accountRepo: ar,
		txManager:   tm,
	}
}

func (uc *TransactionUseCase) CreateManualTransaction(
	ctx context.Context,
	userID uuid.UUID,
	accountID uuid.UUID,
	categoryID *uuid.UUID,
	name string,
	isIncome bool,
	amount int64,
	completedAt time.Time,
	comment *string,
) error {
	trans, err := domain.NewTransaction(
		userID, accountID, categoryID, name, isIncome, amount, completedAt, false, comment,
	)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	delta := amount
	if !isIncome {
		delta = -amount
	}

	return uc.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := uc.transRepo.AddTransaction(ctx, trans); err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}

		if err := uc.accountRepo.UpdateBalance(ctx, userID, accountID, delta); err != nil {
			return fmt.Errorf("transaction created but failed to update account balance: %w", err)
		}

		return nil
	})
}

func (uc *TransactionUseCase) DeleteManualTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	return uc.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		trans, err := uc.transRepo.GetTransaction(ctx, userID, transactionID)
		if err != nil {
			return fmt.Errorf("transaction not found: %w", err)
		}

		if trans.IsImported {
			return ErrCannotModifyImported
		}

		if err := uc.transRepo.DeleteTransaction(ctx, userID, transactionID); err != nil {
			return fmt.Errorf("failed to delete transaction: %w", err)
		}

		if !trans.IsHidden {
			delta := trans.Amount
			if trans.IsIncome {
				delta = -trans.Amount
			}
			if err := uc.accountRepo.UpdateBalance(ctx, userID, trans.AccountID, delta); err != nil {
				return fmt.Errorf("transaction deleted but failed to restore balance: %w", err)
			}
		}

		return nil
	})
}

func (uc *TransactionUseCase) GetUserTransactions(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrTransEmptyUserID
	}

	transactions, err := uc.transRepo.GetAllTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return transactions, nil
}

func (uc *TransactionUseCase) ToggleTransactionsVisibility(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID, hide bool) error {
	if len(transactionIDs) == 0 {
		return nil
	}

	transactions, err := uc.transRepo.GetTransactionsByIDs(ctx, userID, transactionIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	accountDeltas := make(map[uuid.UUID]int64)
	var idsToUpdate []uuid.UUID

	for _, t := range transactions {
		if t.IsHidden == hide {
			continue
		}

		idsToUpdate = append(idsToUpdate, t.TransactionID)
		var delta int64

		if hide {
			if t.IsIncome {
				delta = -t.Amount
			} else {
				delta = t.Amount
			}
		} else {
			if t.IsIncome {
				delta = t.Amount
			} else {
				delta = -t.Amount
			}
		}

		accountDeltas[t.AccountID] += delta
	}

	if len(idsToUpdate) == 0 {
		return nil
	}

	return uc.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		if hide {
			err = uc.transRepo.HideTransactions(ctx, userID, idsToUpdate)
		} else {
			err = uc.transRepo.ShowTransactions(ctx, userID, idsToUpdate)
		}
		if err != nil {
			return fmt.Errorf("failed to toggle visibility in DB: %w", err)
		}

		for accountID, delta := range accountDeltas {
			if delta == 0 {
				continue
			}
			if err := uc.accountRepo.UpdateBalance(ctx, userID, accountID, delta); err != nil {
				return fmt.Errorf("failed to update balance for account %s: %w", accountID, err)
			}
		}

		return nil
	})
}
