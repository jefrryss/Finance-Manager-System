package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/modules/account/usecase"
)

type AccountRouter struct {
	accountUC *usecase.AccountUseCase
}

func NewAccountRouter(accountUC *usecase.AccountUseCase) *AccountRouter {
	return &AccountRouter{
		accountUC: accountUC,
	}
}

func (a *AccountRouter) Route() chi.Router {
	r := chi.NewRouter()

	r.Post("/", a.CreateAccount)
	r.Post("/import/pdf", a.ImportAccountFromPDF)
	r.Get("/", a.GetAccounts)
	r.Put("/{id}", a.RenameAccount)
	r.Delete("/{id}", a.ArchiveAccount)

	return r
}

type CreateAccountReq struct {
	Name              string  `json:"name" example:"Мой кошелек"`
	Currency          string  `json:"currency" example:"RUB"`
	AccountType       string  `json:"account_type" example:"manual"`
	ColorHex          string  `json:"color_hex" example:"#FF0000"`
	IsImported        bool    `json:"is_imported" example:"false"`
	ExternalAccountID *string `json:"external_account_id" example:"null"`
}

type RenameAccountReq struct {
	Name string `json:"name" example:"Новое название кошелька"`
}

// @Summary Создать счет
// @Tags accounts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body CreateAccountReq true "Данные счета"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/accounts [post]
func (a *AccountRouter) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateAccountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = a.accountUC.CreateAccount(
		r.Context(),
		userID,
		req.Name,
		req.Currency,
		req.AccountType,
		req.ColorHex,
		req.IsImported,
		req.ExternalAccountID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Account created",
	})
}

// @Summary Импортировать счет и транзакции из PDF выписки Т-Банка
// @Tags accounts
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF выписка Т-Банка"
// @Param name formData string false "Название счета"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/accounts/import/pdf [post]
func (a *AccountRouter) ImportAccountFromPDF(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}

	accountName := r.FormValue("name")
	result, err := a.accountUC.ImportAccountFromTBankPDF(r.Context(), userID, accountName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":                "success",
		"account_id":            result.AccountID,
		"imported_transactions": result.ImportedTransactions,
		"balance":               result.Balance,
		"account_number":        result.AccountNumber,
		"contract_number":       result.ContractNumber,
	})
}

// @Summary Получить все активные счета
// @Tags accounts
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} domain.Account
// @Router /api/v1/accounts [get]
func (a *AccountRouter) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := a.accountUC.GetUserAccounts(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// @Summary Переименовать счет
// @Tags accounts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID счета"
// @Param request body RenameAccountReq true "Новое название"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/accounts/{id} [put]
func (a *AccountRouter) RenameAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req RenameAccountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = a.accountUC.RenameAccount(r.Context(), userID, accountID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Account renamed",
	})
}

// @Summary Архивировать (удалить) счет
// @Tags accounts
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID счета"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/accounts/{id} [delete]
func (a *AccountRouter) ArchiveAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	err = a.accountUC.ArchiveAccount(r.Context(), userID, accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Account archived",
	})
}
