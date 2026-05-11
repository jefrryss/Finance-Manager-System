package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"

	accountDomain "Finance-Manager-System/internal/infrastructure/modules/account/domain"
	categoryDomain "Finance-Manager-System/internal/infrastructure/modules/category/domain"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type fakeAccountRepo struct {
	account *accountDomain.Account
	updated bool
}

func (f *fakeAccountRepo) AddAccount(ctx context.Context, acc *accountDomain.Account) (uuid.UUID, error) {
	id := uuid.New()
	acc.AccountID = id
	f.account = acc
	return id, nil
}
func (f *fakeAccountRepo) ArchiveAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error {
	return nil
}
func (f *fakeAccountRepo) GetAllAccountsByUser(ctx context.Context, userID uuid.UUID) ([]accountDomain.Account, error) {
	if f.account == nil {
		return nil, nil
	}
	return []accountDomain.Account{*f.account}, nil
}
func (f *fakeAccountRepo) GetAccountByID(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (*accountDomain.Account, error) {
	return f.account, nil
}
func (f *fakeAccountRepo) UpdateAccountName(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, name string) error {
	f.account.NameAccount = name
	f.updated = true
	return nil
}
func (f *fakeAccountRepo) UpdateManualAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, name string, balance int64) error {
	f.account.NameAccount = name
	f.account.Balance = balance
	f.updated = true
	return nil
}
func (f *fakeAccountRepo) UpdateImportedAccountSnapshot(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, balance int64) error {
	f.account.Balance = balance
	return nil
}

type fakeAccountCatRepo struct{}

func (f *fakeAccountCatRepo) GetCategoriesByUser(ctx context.Context, userID uuid.UUID) ([]categoryDomain.Category, error) {
	return nil, nil
}

type fakeAccountTransRepo struct{}

func (f *fakeAccountTransRepo) AddTransactions(ctx context.Context, transactions []*transactionDomain.Transaction) (int, error) {
	return len(transactions), nil
}
func (f *fakeAccountTransRepo) ResolveAutoCategoryID(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string) (*uuid.UUID, error) {
	return nil, nil
}

type fakeTxManager struct{}

func (f *fakeTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestImportAccountFromInvalidPDF(t *testing.T) {
	uc := NewAccountUseCase(&fakeAccountRepo{}, &fakeAccountCatRepo{}, &fakeAccountTransRepo{}, &fakeTxManager{})
	_, err := uc.ImportAccountFromTBankPDF(context.Background(), uuid.New(), "x", []byte("not pdf"))
	if err != ErrInvalidStatement {
		t.Fatalf("expected ErrInvalidStatement, got %v", err)
	}
}

func TestUpdateManualAccountRejectsBalanceForImported(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	ext := "acc"
	repo := &fakeAccountRepo{
		account: &accountDomain.Account{
			AccountID:         accountID,
			UserID:            userID,
			IsImported:        true,
			ExternalAccountID: &ext,
			NameAccount:       "Imported",
			Balance:           100,
		},
	}
	uc := NewAccountUseCase(repo, &fakeAccountCatRepo{}, &fakeAccountTransRepo{}, &fakeTxManager{})
	nextBalance := int64(200)
	err := uc.UpdateManualAccount(context.Background(), userID, accountID, "Renamed", &nextBalance)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUpdateManualAccountSuccess(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	repo := &fakeAccountRepo{
		account: &accountDomain.Account{
			AccountID:   accountID,
			UserID:      userID,
			IsImported:  false,
			NameAccount: "Manual",
			Balance:     100,
		},
	}
	uc := NewAccountUseCase(repo, &fakeAccountCatRepo{}, &fakeAccountTransRepo{}, &fakeTxManager{})
	nextBalance := int64(333)
	err := uc.UpdateManualAccount(context.Background(), userID, accountID, "Manual 2", &nextBalance)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.account.NameAccount != "Manual 2" || repo.account.Balance != 333 {
		t.Fatalf("unexpected account data")
	}
}

