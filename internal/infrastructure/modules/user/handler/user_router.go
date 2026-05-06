package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"Finance-Manager-System/internal/infrastructure/modules/user/usecase"
)

type UserRouter struct {
	userCase *usecase.UserCase
}

func NewUserRouter(userCase *usecase.UserCase) *UserRouter {
	return &UserRouter{
		userCase: userCase,
	}
}

func (u *UserRouter) Route() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", u.Register)
	r.Put("/change_password", u.ChangePassword)

	return r
}

type RegisterReq struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ChangePasswordReq struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

// @Summary Регистрация пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterReq true "Данные для регистрации"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/users/register [post]
func (u *UserRouter) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	id, err := u.userCase.RegistrateUser(r.Context(), req.Email, req.Login, req.Password)
	if err != nil {
		u.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "User created",
		"id":      id,
	})
}

// @Summary Изменить пароль
// @Tags users
// @Accept json
// @Produce json
// @Param request body ChangePasswordReq true "Данные для смены пароля"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/users/change_password [put]
func (u *UserRouter) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		http.Error(w, "Invalid uuid", http.StatusBadRequest)
		return
	}

	err = u.userCase.ChangeUserPassword(r.Context(), id, req.Password)
	if err != nil {
		u.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Password changed",
	})
}

func (u *UserRouter) mapError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		statusCode = http.StatusConflict
		message = "User already exists"
	case errors.Is(err, domain.ErrInvalidEmail):
		statusCode = http.StatusBadRequest
		message = "Invalid input email"
	case errors.Is(err, domain.ErrInvalidLogin):
		statusCode = http.StatusBadRequest
		message = "Invalid input login"
	case errors.Is(err, domain.ErrInvalidPassword):
		statusCode = http.StatusBadRequest
		message = "Invalid input password"
	default:
		log.Printf("DB error: %v", err)
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}
	http.Error(w, message, statusCode)
}
