package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/middleware"
	accountDomain "Finance-Manager-System/internal/infrastructure/modules/account/domain"
	accountUsecase "Finance-Manager-System/internal/infrastructure/modules/account/usecase"
	categoryDomain "Finance-Manager-System/internal/infrastructure/modules/category/domain"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type integrationAccountRepo struct {
	items map[uuid.UUID]*accountDomain.Account
}

func newIntegrationAccountRepo() *integrationAccountRepo {
	return &integrationAccountRepo{items: make(map[uuid.UUID]*accountDomain.Account)}
}

func (r *integrationAccountRepo) AddAccount(ctx context.Context, acc *accountDomain.Account) (uuid.UUID, error) {
	if acc.AccountID == uuid.Nil {
		acc.AccountID = uuid.New()
	}
	r.items[acc.AccountID] = acc
	return acc.AccountID, nil
}
func (r *integrationAccountRepo) ArchiveAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error {
	return nil
}
func (r *integrationAccountRepo) GetAllAccountsByUser(ctx context.Context, userID uuid.UUID) ([]accountDomain.Account, error) {
	out := make([]accountDomain.Account, 0)
	for _, v := range r.items {
		if v.UserID == userID {
			out = append(out, *v)
		}
	}
	return out, nil
}
func (r *integrationAccountRepo) GetAccountByID(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (*accountDomain.Account, error) {
	return r.items[accountID], nil
}
func (r *integrationAccountRepo) UpdateAccountName(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, name string) error {
	r.items[accountID].NameAccount = name
	return nil
}
func (r *integrationAccountRepo) UpdateManualAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, name string, balance int64) error {
	r.items[accountID].NameAccount = name
	r.items[accountID].Balance = balance
	return nil
}
func (r *integrationAccountRepo) UpdateImportedAccountSnapshot(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, balance int64) error {
	r.items[accountID].Balance = balance
	return nil
}

type integrationAccountCategoryRepo struct{}

func (r *integrationAccountCategoryRepo) GetCategoriesByUser(ctx context.Context, userID uuid.UUID) ([]categoryDomain.Category, error) {
	return nil, nil
}

type integrationAccountTransRepo struct{}

func (r *integrationAccountTransRepo) AddTransactions(ctx context.Context, transactions []*transactionDomain.Transaction) (int, error) {
	return len(transactions), nil
}
func (r *integrationAccountTransRepo) ResolveAutoCategoryID(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string) (*uuid.UUID, error) {
	return nil, nil
}

type integrationAccountTxManager struct{}

func (m *integrationAccountTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func withAccountUser(req *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestAccountRouterCreateManual(t *testing.T) {
	repo := newIntegrationAccountRepo()
	uc := accountUsecase.NewAccountUseCase(repo, &integrationAccountCategoryRepo{}, &integrationAccountTransRepo{}, &integrationAccountTxManager{})
	router := NewAccountRouter(uc).Route()
	userID := uuid.New()
	body := map[string]interface{}{
		"name":            "Manual",
		"currency":        "RUB",
		"account_type":    "manual",
		"color_hex":       "#000000",
		"is_imported":     false,
		"initial_balance": int64(1000),
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, withAccountUser(req, userID))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAccountRouterImportInvalidPDF(t *testing.T) {
	repo := newIntegrationAccountRepo()
	uc := accountUsecase.NewAccountUseCase(repo, &integrationAccountCategoryRepo{}, &integrationAccountTransRepo{}, &integrationAccountTxManager{})
	router := NewAccountRouter(uc).Route()
	userID := uuid.New()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "bad.pdf")
	part.Write([]byte("bad-pdf-data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/import/pdf", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, withAccountUser(req, userID))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
}

