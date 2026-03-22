package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
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
	r.Put("/{id}", t.UpdateTransaction)
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

type UpdateTransReq struct {
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
		t.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

// @Summary Получить транзакции (с фильтрами)
// @Tags transactions
// @Produce json
// @Param X-User-ID header string true "ID пользователя"
// @Param account_id query string false "ID счета"
// @Param category_id query string false "ID категории"
// @Param is_income query boolean false "Тип (доход/расход)"
// @Param start_date query string false "Начальная дата (RFC3339)"
// @Param end_date query string false "Конечная дата (RFC3339)"
// @Param is_hidden query boolean false "Показать скрытые"
// @Success 200 {array} domain.Transaction
// @Router /api/v1/transactions [get]
func (t *TransactionRouter) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	var filter domain.TransactionFilter

	if accID := r.URL.Query().Get("account_id"); accID != "" {
		if id, err := uuid.Parse(accID); err == nil {
			filter.AccountID = &id
		}
	}
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
		}
	}
	if isInc := r.URL.Query().Get("is_income"); isInc != "" {
		val := isInc == "true"
		filter.IsIncome = &val
	}
	if isHid := r.URL.Query().Get("is_hidden"); isHid != "" {
		val := isHid == "true"
		filter.IsHidden = &val
	}
	if start := r.URL.Query().Get("start_date"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartDate = &parsed
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndDate = &parsed
		}
	}

	transactions, err := t.transUC.GetUserTransactions(r.Context(), userID, filter)
	if err != nil {
		t.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// @Summary Обновить транзакцию
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-User-ID header string true "ID пользователя"
// @Param id path string true "ID транзакции"
// @Param request body UpdateTransReq true "Данные для обновления"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/transactions/{id} [put]
func (t *TransactionRouter) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateTransReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = t.transUC.UpdateTransaction(
		r.Context(), userID, transID, req.CategoryID,
		req.Name, req.IsIncome, req.Amount, req.CompletedAt, req.Comment,
	)

	if err != nil {
		t.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
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
		t.mapError(w, err)
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
		t.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func (t *TransactionRouter) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTransNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrCannotModifyImported):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrTransInvalidAmount),
		errors.Is(err, domain.ErrTransEmptyName),
		errors.Is(err, domain.ErrTransEmptyAccountID):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
