package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type TransactionRepository interface {
	GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*domain.Transaction, error)
	AddTransaction(ctx context.Context, trans *domain.Transaction) error
	UpdateTransaction(ctx context.Context, trans *domain.Transaction) error
	DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error
	GetAllTransactions(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error)
	GetTransactionsWithFilter(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, error)
	ShowTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error
	HideTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error
	GetTransactionsByIDs(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID) ([]domain.Transaction, error)
	ResolveAutoCategoryID(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string) (*uuid.UUID, error)
	UpsertAutoCategoryRule(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string, categoryID uuid.UUID) error
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

func (uc *TransactionUseCase) CreateManualTransaction(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, categoryID *uuid.UUID, name string, isIncome bool, amount int64, completedAt time.Time, comment *string, currency string, bankFee int64, status string) error {
	trans, err := domain.NewTransaction(userID, accountID, categoryID, name, isIncome, amount, completedAt, false, comment)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if currency != "" {
		trans.Currency = currency
	}
	if status != "" {
		trans.Status = status
	}
	if bankFee > 0 {
		trans.BankFee = bankFee
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

func (uc *TransactionUseCase) UpdateTransaction(ctx context.Context, userID, transID uuid.UUID, categoryID *uuid.UUID, name string, isIncome bool, amount int64, completedAt time.Time, comment *string, currency string, bankFee int64, status string) error {
	return uc.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		oldTrans, err := uc.transRepo.GetTransaction(ctx, userID, transID)
		if err != nil {
			return fmt.Errorf("failed to fetch transaction: %w", err)
		}

		if oldTrans.IsImported {
			nextCurrency := oldTrans.Currency
			if currency != "" {
				nextCurrency = currency
			}
			nextStatus := oldTrans.Status
			if status != "" {
				nextStatus = status
			}
			nextBankFee := oldTrans.BankFee
			if bankFee >= 0 {
				nextBankFee = bankFee
			}
			if amount != oldTrans.Amount || isIncome != oldTrans.IsIncome || !completedAt.Equal(oldTrans.CompletedAt) || name != oldTrans.NameTransaction || nextCurrency != oldTrans.Currency || nextStatus != oldTrans.Status || nextBankFee != oldTrans.BankFee {
				return domain.ErrCannotModifyImported
			}
		} else {
			if amount <= 0 {
				return domain.ErrTransInvalidAmount
			}
		}

		var balanceDelta int64 = 0
		if !oldTrans.IsImported && !oldTrans.IsHidden {
			oldDelta := oldTrans.Amount
			if !oldTrans.IsIncome {
				oldDelta = -oldTrans.Amount
			}

			newDelta := amount
			if !isIncome {
				newDelta = -amount
			}

			balanceDelta = newDelta - oldDelta
		}

		oldTrans.CategoryID = categoryID
		oldTrans.NameTransaction = name
		oldTrans.IsIncome = isIncome
		oldTrans.Amount = amount
		oldTrans.CompletedAt = completedAt
		oldTrans.Comment = comment
		if currency != "" {
			oldTrans.Currency = currency
		}
		if status != "" {
			oldTrans.Status = status
		}
		if bankFee >= 0 {
			oldTrans.BankFee = bankFee
		}

		if err := uc.transRepo.UpdateTransaction(ctx, oldTrans); err != nil {
			return fmt.Errorf("failed to update transaction in db: %w", err)
		}
		if oldTrans.IsImported && categoryID != nil {
			if upsertErr := uc.transRepo.UpsertAutoCategoryRule(ctx, userID, oldTrans.IsIncome, oldTrans.MCCCode, oldTrans.NameTransaction, *categoryID); upsertErr != nil {
				return fmt.Errorf("failed to save auto-category rule: %w", upsertErr)
			}
		}

		if balanceDelta != 0 {
			if err := uc.accountRepo.UpdateBalance(ctx, userID, oldTrans.AccountID, balanceDelta); err != nil {
				return fmt.Errorf("failed to update account balance during update: %w", err)
			}
		}
		return nil
	})
}

func (uc *TransactionUseCase) UpdateImportedTransactionMeta(
	ctx context.Context,
	userID uuid.UUID,
	transID uuid.UUID,
	categoryID *uuid.UUID,
	comment *string,
	isHidden *bool,
) error {
	return uc.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		trans, err := uc.transRepo.GetTransaction(txCtx, userID, transID)
		if err != nil {
			return fmt.Errorf("failed to fetch transaction: %w", err)
		}
		if !trans.IsImported {
			return domain.ErrCannotModifyImported
		}

		if categoryID != nil {
			trans.CategoryID = categoryID
			if upsertErr := uc.transRepo.UpsertAutoCategoryRule(txCtx, userID, trans.IsIncome, trans.MCCCode, trans.NameTransaction, *categoryID); upsertErr != nil {
				return fmt.Errorf("failed to save auto-category rule: %w", upsertErr)
			}
		}
		if comment != nil {
			cleaned := strings.TrimSpace(*comment)
			if cleaned == "" {
				trans.Comment = nil
			} else {
				trans.Comment = &cleaned
			}
		}

		if isHidden != nil && trans.IsHidden != *isHidden {
			var delta int64
			if *isHidden {
				if trans.IsIncome {
					delta = -trans.Amount
				} else {
					delta = trans.Amount
				}
			} else {
				if trans.IsIncome {
					delta = trans.Amount
				} else {
					delta = -trans.Amount
				}
			}
			if delta != 0 {
				if err := uc.accountRepo.UpdateBalance(txCtx, userID, trans.AccountID, delta); err != nil {
					return fmt.Errorf("failed to update account balance: %w", err)
				}
			}
			trans.IsHidden = *isHidden
		}

		if err := uc.transRepo.UpdateTransaction(txCtx, trans); err != nil {
			return fmt.Errorf("failed to update imported transaction: %w", err)
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
			return domain.ErrCannotModifyImported
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

func (uc *TransactionUseCase) GetUserTransactions(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrTransEmptyUserID
	}

	transactions, err := uc.transRepo.GetTransactionsWithFilter(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	sortTransactionsDesc(transactions)
	return transactions, nil
}

func sortTransactionsDesc(transactions []domain.Transaction) {
	sort.SliceStable(transactions, func(i, j int) bool {
		left := transactions[i]
		right := transactions[j]
		if !left.CompletedAt.Equal(right.CompletedAt) {
			return left.CompletedAt.After(right.CompletedAt)
		}
		return left.TransactionID.String() > right.TransactionID.String()
	})
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
