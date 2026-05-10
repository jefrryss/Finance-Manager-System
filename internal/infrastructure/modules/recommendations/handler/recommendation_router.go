package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/modules/recommendations/usecase"
)

type RecommendationRouter struct {
	uc *usecase.RecommendationUseCase
}

func NewRecommendationRouter(uc *usecase.RecommendationUseCase) *RecommendationRouter {
	return &RecommendationRouter{uc: uc}
}

func (h *RecommendationRouter) Route() chi.Router {
	r := chi.NewRouter()
	r.Get("/budget", h.GetBudgetRecommendations)
	return r
}

func (h *RecommendationRouter) GetBudgetRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	plannedTotalStr := r.URL.Query().Get("planned_total")
	if plannedTotalStr == "" {
		http.Error(w, "planned_total is required", http.StatusBadRequest)
		return
	}
	plannedTotal, err := strconv.ParseInt(plannedTotalStr, 10, 64)
	if err != nil {
		http.Error(w, "planned_total must be an integer", http.StatusBadRequest)
		return
	}

	months := 3
	if monthsStr := r.URL.Query().Get("months"); monthsStr != "" {
		parsedMonths, parseErr := strconv.Atoi(monthsStr)
		if parseErr != nil {
			http.Error(w, "months must be an integer", http.StatusBadRequest)
			return
		}
		months = parsedMonths
	}

	includeHidden := false
	if includeHiddenStr := r.URL.Query().Get("include_hidden"); includeHiddenStr != "" {
		includeHidden = includeHiddenStr == "true"
	}

	accountIDs, err := parseAccountIDs(r.URL.Query().Get("account_ids"))
	if err != nil {
		http.Error(w, "account_ids must contain valid UUIDs", http.StatusBadRequest)
		return
	}

	result, err := h.uc.GetBudgetRecommendations(r.Context(), userID, plannedTotal, months, includeHidden, accountIDs)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidPlannedTotal), errors.Is(err, usecase.ErrInvalidMonths):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"planned_total":   plannedTotal,
		"months":          months,
		"include_hidden":  includeHidden,
		"account_ids":     accountIDs,
		"recommendations": result,
	})
}

func parseAccountIDs(raw string) ([]uuid.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	ids := make([]uuid.UUID, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := uuid.Parse(part)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
