package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/middleware"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
	transactionUsecase "Finance-Manager-System/internal/infrastructure/modules/transactions/usecase"
)

type integrationTransRepo struct {
	items map[uuid.UUID]*transactionDomain.Transaction
}

func newIntegrationTransRepo() *integrationTransRepo {
	return &integrationTransRepo{items: make(map[uuid.UUID]*transactionDomain.Transaction)}
}

func (r *integrationTransRepo) GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*transactionDomain.Transaction, error) {
	item, ok := r.items[transactionID]
	if !ok {
		return nil, transactionDomain.ErrTransNotFound
	}
	return item, nil
}
func (r *integrationTransRepo) AddTransaction(ctx context.Context, trans *transactionDomain.Transaction) error {
	if trans.TransactionID == uuid.Nil {
		trans.TransactionID = uuid.New()
	}
	r.items[trans.TransactionID] = trans
	return nil
}
func (r *integrationTransRepo) UpdateTransaction(ctx context.Context, trans *transactionDomain.Transaction) error {
	r.items[trans.TransactionID] = trans
	return nil
}
func (r *integrationTransRepo) DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	delete(r.items, transactionID)
	return nil
}
func (r *integrationTransRepo) GetAllTransactions(ctx context.Context, userID uuid.UUID) ([]transactionDomain.Transaction, error) {
	out := make([]transactionDomain.Transaction, 0, len(r.items))
	for _, v := range r.items {
		if v.UserID == userID {
			out = append(out, *v)
		}
	}
	return out, nil
}
func (r *integrationTransRepo) GetTransactionsWithFilter(ctx context.Context, userID uuid.UUID, filter transactionDomain.TransactionFilter) ([]transactionDomain.Transaction, error) {
	out := make([]transactionDomain.Transaction, 0, len(r.items))
	for _, v := range r.items {
		if v.UserID != userID {
			continue
		}
		out = append(out, *v)
	}
	return out, nil
}
func (r *integrationTransRepo) ShowTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error {
	for _, id := range transactionIds {
		if tx, ok := r.items[id]; ok {
			tx.IsHidden = false
		}
	}
	return nil
}
func (r *integrationTransRepo) HideTransactions(ctx context.Context, userID uuid.UUID, transactionIds []uuid.UUID) error {
	for _, id := range transactionIds {
		if tx, ok := r.items[id]; ok {
			tx.IsHidden = true
		}
	}
	return nil
}
func (r *integrationTransRepo) GetTransactionsByIDs(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID) ([]transactionDomain.Transaction, error) {
	out := make([]transactionDomain.Transaction, 0, len(transactionIDs))
	for _, id := range transactionIDs {
		if tx, ok := r.items[id]; ok {
			out = append(out, *tx)
		}
	}
	return out, nil
}
func (r *integrationTransRepo) ResolveAutoCategoryID(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string) (*uuid.UUID, error) {
	return nil, nil
}
func (r *integrationTransRepo) UpsertAutoCategoryRule(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string, categoryID uuid.UUID) error {
	return nil
}

type integrationBalanceRepo struct{}

func (r *integrationBalanceRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, amountDelta int64) error {
	return nil
}

type integrationTxManager struct{}

func (m *integrationTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func withUser(req *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestTransactionRouterCreateAndGet(t *testing.T) {
	repo := newIntegrationTransRepo()
	uc := transactionUsecase.NewTransactionUseCase(repo, &integrationBalanceRepo{}, &integrationTxManager{})
	router := NewTransactionRouter(uc).Route()
	userID := uuid.New()
	accountID := uuid.New()

	createBody := map[string]interface{}{
		"account_id":   accountID,
		"name":         "Lunch",
		"is_income":    false,
		"amount":       int64(15000),
		"completed_at": time.Now().UTC().Format(time.RFC3339),
		"currency":     "RUB",
		"bank_fee":     int64(0),
		"status":       "completed",
	}
	b, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(req, userID))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
	rrGet := httptest.NewRecorder()
	router.ServeHTTP(rrGet, withUser(reqGet, userID))
	if rrGet.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", rrGet.Code, rrGet.Body.String())
	}
}

func TestTransactionRouterPatchImported(t *testing.T) {
	repo := newIntegrationTransRepo()
	uc := transactionUsecase.NewTransactionUseCase(repo, &integrationBalanceRepo{}, &integrationTxManager{})
	router := NewTransactionRouter(uc).Route()
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	repo.items[txID] = &transactionDomain.Transaction{
		TransactionID:   txID,
		UserID:          userID,
		AccountID:       accountID,
		NameTransaction: "Store",
		IsIncome:        false,
		Amount:          10000,
		CompletedAt:     time.Now().UTC(),
		IsImported:      true,
		Currency:        "RUB",
		Status:          "completed",
	}

	categoryID := uuid.New()
	hidden := true
	comment := "checked"
	patchBody := map[string]interface{}{
		"category_id": categoryID,
		"is_hidden":   hidden,
		"comment":     comment,
	}
	b, _ := json.Marshal(patchBody)
	req := httptest.NewRequest(http.MethodPatch, "/"+txID.String()+"/imported", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(req, userID))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}

	updated := repo.items[txID]
	if updated.CategoryID == nil || *updated.CategoryID != categoryID {
		t.Fatalf("category not updated")
	}
	if updated.Comment == nil || *updated.Comment != "checked" {
		t.Fatalf("comment not updated")
	}
	if !updated.IsHidden {
		t.Fatalf("hidden not updated")
	}
}

