package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/transactions/usecase"
)

type TransactionRouter struct {
	transUC *usecase.TransactionUseCase
}

func NewTransactionRouter(transUC *usecase.TransactionUseCase) *TransactionRouter {
	return &TransactionRouter{
		transUC: transUC,
	}
}

func (t *TransactionRouter) Route() chi.Router {
	r := chi.NewRouter()

	r.Post("/", t.CreateTransaction)
	r.Get("/", t.GetTransactions)
	r.Delete("/{id}", t.DeleteTransaction)
	r.Patch("/visibility", t.ToggleVisibility)

	return r
}

type CreateTransReq struct {
	AccountID   uuid.UUID  `json:"account_id"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Name        string     `json:"name"`
	IsIncome    bool       `json:"is_income"`
	Amount      int64      `json:"amount"`
	CompletedAt time.Time  `json:"completed_at"`
	Comment     *string    `json:"comment"`
}

type ToggleVisibilityReq struct {
	TransactionIDs []uuid.UUID `json:"transaction_ids"`
	Hide           bool        `json:"hide"`
}

// @Summary Создать транзакцию
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-User-ID header string true "ID пользователя"
// @Param request body CreateTransReq true "Данные транзакции"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/transactions [post]
func (t *TransactionRouter) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	var req CreateTransReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = t.transUC.CreateManualTransaction(
		r.Context(), userID, req.AccountID, req.CategoryID,
		req.Name, req.IsIncome, req.Amount, req.CompletedAt, req.Comment,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

// @Summary Получить транзакции
// @Tags transactions
// @Produce json
// @Param X-User-ID header string true "ID пользователя"
// @Success 200 {array} domain.Transaction
// @Router /api/v1/transactions [get]
func (t *TransactionRouter) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	transactions, err := t.transUC.GetUserTransactions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// @Summary Удалить транзакцию
// @Tags transactions
// @Produce json
// @Param X-User-ID header string true "ID пользователя"
// @Param id path string true "ID транзакции"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/transactions/{id} [delete]
func (t *TransactionRouter) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	transID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	err = t.transUC.DeleteManualTransaction(r.Context(), userID, transID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

// @Summary Изменить видимость транзакций
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-User-ID header string true "ID пользователя"
// @Param request body ToggleVisibilityReq true "IDs транзакций и статус"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/transactions/visibility [patch]
func (t *TransactionRouter) ToggleVisibility(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	var req ToggleVisibilityReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = t.transUC.ToggleTransactionsVisibility(r.Context(), userID, req.TransactionIDs, req.Hide)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
