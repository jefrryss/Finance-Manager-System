package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/modules/goals/domain"
	"Finance-Manager-System/internal/infrastructure/modules/goals/usecase"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type GoalRouter struct {
	goalUC *usecase.GoalUseCase
}

func NewGoalRouter(goalUC *usecase.GoalUseCase) *GoalRouter {
	return &GoalRouter{goalUC: goalUC}
}

func (h *GoalRouter) Route() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateGoal)
	r.Get("/", h.GetGoals)
	r.Get("/{id}", h.GetGoalDetails)
	r.Put("/{id}", h.UpdateGoal)
	r.Delete("/{id}", h.DeleteGoal)
	r.Post("/{id}/contributions", h.AddContribution)
	return r
}

type CreateGoalReq struct {
	NameGoal     string     `json:"name_goal"`
	TargetAmount int64      `json:"target_amount"`
	TargetDate   *time.Time `json:"target_date"`
}

type UpdateGoalReq struct {
	NameGoal     string     `json:"name_goal"`
	TargetAmount int64      `json:"target_amount"`
	TargetDate   *time.Time `json:"target_date"`
}

type AddContributionReq struct {
	Amount           int64      `json:"amount"`
	ContributionDate *time.Time `json:"contribution_date"`
	TransactionID    *uuid.UUID `json:"transaction_id"`
}

// @Summary Создать новую цель
// @Tags goals
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body CreateGoalReq true "Данные цели"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/goals [post]
func (h *GoalRouter) CreateGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateGoalReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	goalID, err := h.goalUC.CreateGoal(r.Context(), userID, req.NameGoal, req.TargetAmount, req.TargetDate)
	if err != nil {
		h.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "goal_id": goalID})
}

// @Summary Получить список целей (сводка)
// @Tags goals
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} domain.GoalSummary
// @Router /api/v1/goals [get]
func (h *GoalRouter) GetGoals(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goals, err := h.goalUC.GetGoals(r.Context(), userID)
	if err != nil {
		h.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

// @Summary Получить детали цели (история и прогноз)
// @Tags goals
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID цели"
// @Success 200 {object} domain.GoalDetails
// @Router /api/v1/goals/{id} [get]
func (h *GoalRouter) GetGoalDetails(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	details, err := h.goalUC.GetGoalDetails(r.Context(), userID, goalID)
	if err != nil {
		h.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// @Summary Обновить параметры цели
// @Tags goals
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID цели"
// @Param request body UpdateGoalReq true "Новые данные цели"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/goals/{id} [put]
func (h *GoalRouter) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	var req UpdateGoalReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if err := h.goalUC.UpdateGoal(r.Context(), userID, goalID, req.NameGoal, req.TargetAmount, req.TargetDate); err != nil {
		h.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

// @Summary Удалить цель
// @Tags goals
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID цели"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/goals/{id} [delete]
func (h *GoalRouter) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	if err := h.goalUC.DeleteGoal(r.Context(), userID, goalID); err != nil {
		h.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

// @Summary Добавить пополнение цели
// @Tags goals
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID цели"
// @Param request body AddContributionReq true "Данные пополнения"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/goals/{id}/contributions [post]
func (h *GoalRouter) AddContribution(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	var req AddContributionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	contributionID, err := h.goalUC.AddContribution(r.Context(), userID, goalID, req.Amount, req.ContributionDate, req.TransactionID)
	if err != nil {
		h.mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "contribution_id": contributionID})
}

func (h *GoalRouter) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrGoalNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, transactionDomain.ErrTransNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrGoalEmptyName),
		errors.Is(err, domain.ErrGoalNameTooLong),
		errors.Is(err, domain.ErrGoalInvalidTargetAmount),
		errors.Is(err, domain.ErrGoalInvalidContributionAmount),
		errors.Is(err, domain.ErrGoalEmptyUserID):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
