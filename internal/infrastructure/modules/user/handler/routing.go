package handler

import (
	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"Finance-Manager-System/internal/infrastructure/modules/user/usecase"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Password string `json:"password"`
			Email    string `json:"email"`
			Login    string `json:"login"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid json", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		id, err := u.userCase.RegistrateUser(ctx, req.Email, req.Login, req.Password)
		if err != nil {
			u.mapError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		response := map[string]interface{}{
			"status":  "success",
			"message": "User created",
			"id":      id,
		}

		json.NewEncoder(w).Encode(response)
	})

	r.Put("/change_password", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Id       string `json:"id"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid json", http.StatusBadGateway)
			return
		}

		ctx := r.Context()

		id, err := uuid.Parse(req.Id)
		if err != nil {
			http.Error(w, "Invalid uuid", http.StatusBadRequest)
			return
		}
		err = u.userCase.ChangeUserPassword(ctx, id, req.Password)

		if err != nil {
			u.mapError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		response := map[string]interface{}{
			"status":  "success",
			"message": "Password changed",
		}
		json.NewEncoder(w).Encode(response)
	})

	return r
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
		log.Printf("Ошибка бд %v", err)
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}
	http.Error(w, message, statusCode)

}
