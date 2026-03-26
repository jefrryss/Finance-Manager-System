package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/modules/analytics/usecase"
)

type AnalyticsRouter struct {
	analyticsUC *usecase.AnalyticsUseCase
}

func NewAnalyticsRouter(analyticsUC *usecase.AnalyticsUseCase) *AnalyticsRouter {
	return &AnalyticsRouter{analyticsUC: analyticsUC}
}

func (a *AnalyticsRouter) Route() chi.Router {
	r := chi.NewRouter()

	r.Get("/summary", a.GetSummary)
	r.Get("/categories", a.GetCategories)
	r.Get("/daily", a.GetDaily)

	return r
}

func parseDates(r *http.Request) (*time.Time, *time.Time) {
	var startPtr, endPtr *time.Time
	if start := r.URL.Query().Get("start_date"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			startPtr = &parsed
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			endPtr = &parsed
		}
	}
	return startPtr, endPtr
}

// @Summary Получить сводку (доходы и расходы)
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param start_date query string false "Начальная дата (RFC3339)"
// @Param end_date query string false "Конечная дата (RFC3339)"
// @Success 200 {object} domain.SummaryReport
// @Router /api/v1/analytics/summary [get]
func (a *AnalyticsRouter) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	start, end := parseDates(r)
	report, err := a.analyticsUC.GetSummary(r.Context(), userID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// @Summary Получить траты по категориям
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param start_date query string false "Начальная дата (RFC3339)"
// @Param end_date query string false "Конечная дата (RFC3339)"
// @Success 200 {array} domain.CategoryReport
// @Router /api/v1/analytics/categories [get]
func (a *AnalyticsRouter) GetCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	start, end := parseDates(r)
	report, err := a.analyticsUC.GetCategoryReport(r.Context(), userID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// @Summary Получить динамику по дням
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param is_income query boolean false "Доходы (true) или расходы (false)"
// @Param start_date query string false "Начальная дата (RFC3339)"
// @Param end_date query string false "Конечная дата (RFC3339)"
// @Success 200 {array} domain.DailyReport
// @Router /api/v1/analytics/daily [get]
func (a *AnalyticsRouter) GetDaily(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isIncome := r.URL.Query().Get("is_income") == "true"
	start, end := parseDates(r)

	report, err := a.analyticsUC.GetDailyDynamics(r.Context(), userID, start, end, isIncome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
