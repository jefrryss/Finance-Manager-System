package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type fakeTransRepo struct {
	byID             map[uuid.UUID]*transactionDomain.Transaction
	upsertRulesCount int
}

func (f *fakeTransRepo) GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*transactionDomain.Transaction, error) {
	tx, ok := f.byID[transactionID]
	if !ok {
		return nil, transactionDomain.ErrTransNotFound
	}
	return tx, nil
}
func (f *fakeTransRepo) AddTransaction(ctx context.Context, trans *transactionDomain.Transaction) error {
	if f.byID == nil {
		f.byID = make(map[uuid.UUID]*transactionDomain.Transaction)
	}
	if trans.TransactionID == uuid.Nil {
		trans.TransactionID = uuid.New()
	}
	f.byID[trans.TransactionID] = trans
	return nil
}
func (f *fakeTransRepo) UpdateTransaction(ctx context.Context, trans *transactionDomain.Transaction) error {
	f.byID[trans.TransactionID] = trans
	return nil
}
func (f *fakeTransRepo) DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	delete(f.byID, transactionID)
	return nil
}
func (f *fakeTransRepo) GetAllTransactions(ctx context.Context, userID uuid.UUID) ([]transactionDomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransRepo) GetTransactionsWithFilter(ctx context.Context, userID uuid.UUID, filter transactionDomain.TransactionFilter) ([]transactionDomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransRepo) ShowTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error {
	return nil
}
func (f *fakeTransRepo) HideTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error {
	return nil
}
func (f *fakeTransRepo) GetTransactionsByIDs(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID) ([]transactionDomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransRepo) ResolveAutoCategoryID(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string) (*uuid.UUID, error) {
	return nil, nil
}
func (f *fakeTransRepo) UpsertAutoCategoryRule(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string, categoryID uuid.UUID) error {
	f.upsertRulesCount++
	return nil
}

type fakeBalanceUpdater struct {
	calls []int64
}

func (f *fakeBalanceUpdater) UpdateBalance(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, amountDelta int64) error {
	f.calls = append(f.calls, amountDelta)
	return nil
}

type fakeTransTxManager struct{}

func (f *fakeTransTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestUpdateImportedTransactionMeta(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	repo := &fakeTransRepo{
		byID: map[uuid.UUID]*transactionDomain.Transaction{
			txID: {
				TransactionID:   txID,
				UserID:          userID,
				AccountID:       accountID,
				NameTransaction: "Market",
				IsIncome:        false,
				Amount:          10000,
				CompletedAt:     time.Now().UTC(),
				IsImported:      true,
				IsHidden:        false,
				Currency:        "RUB",
				Status:          "completed",
			},
		},
	}
	balance := &fakeBalanceUpdater{}
	uc := NewTransactionUseCase(repo, balance, &fakeTransTxManager{})
	comment := "manual"
	hide := true
	catID := uuid.New()
	err := uc.UpdateImportedTransactionMeta(context.Background(), userID, txID, &catID, &comment, &hide)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	got := repo.byID[txID]
	if got.CategoryID == nil || *got.CategoryID != catID {
		t.Fatalf("category not updated")
	}
	if got.Comment == nil || *got.Comment != "manual" {
		t.Fatalf("comment not updated")
	}
	if !got.IsHidden {
		t.Fatalf("hidden flag not updated")
	}
	if len(balance.calls) != 1 || balance.calls[0] != 10000 {
		t.Fatalf("unexpected balance updates: %#v", balance.calls)
	}
	if repo.upsertRulesCount != 1 {
		t.Fatalf("expected rule upsert call")
	}
}

func TestUpdateTransactionRejectsImportedCoreFields(t *testing.T) {
	userID := uuid.New()
	txID := uuid.New()
	accountID := uuid.New()
	repo := &fakeTransRepo{
		byID: map[uuid.UUID]*transactionDomain.Transaction{
			txID: {
				TransactionID:   txID,
				UserID:          userID,
				AccountID:       accountID,
				NameTransaction: "Salary",
				IsIncome:        true,
				Amount:          1000,
				CompletedAt:     time.Now().UTC(),
				IsImported:      true,
				Currency:        "RUB",
				Status:          "completed",
			},
		},
	}
	uc := NewTransactionUseCase(repo, &fakeBalanceUpdater{}, &fakeTransTxManager{})
	err := uc.UpdateTransaction(context.Background(), userID, txID, nil, "Salary", true, 2000, repo.byID[txID].CompletedAt, nil, "RUB", 0, "completed")
	if err == nil {
		t.Fatalf("expected error")
	}
}

