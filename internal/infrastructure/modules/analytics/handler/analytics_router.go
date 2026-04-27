package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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
	r.Get("/monthly", a.GetMonthly)
	r.Get("/compare/categories", a.CompareCategories)

	return r
}

func parseDates(r *http.Request) (*time.Time, *time.Time, error) {
	var startPtr, endPtr *time.Time
	if start := r.URL.Query().Get("start_date"); start != "" {
		parsed, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return nil, nil, err
		}
		startPtr = &parsed
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		parsed, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return nil, nil, err
		}
		endPtr = &parsed
	}
	return startPtr, endPtr, nil
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

func parseMonth(value string) (time.Time, error) {
	return time.Parse("2006-01", value)
}

// @Summary Получить сводку (доходы и расходы)
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param start_date query string false "Начальная дата (RFC3339)"
// @Param end_date query string false "Конечная дата (RFC3339)"
// @Param period query string false "Период по умолчанию: day/week/month"
// @Param include_hidden query boolean false "Учитывать скрытые транзакции"
// @Param account_ids query string false "CSV список account_id для фильтра"
// @Success 200 {object} domain.SummaryReport
// @Router /api/v1/analytics/summary [get]
func (a *AnalyticsRouter) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	start, end, err := parseDates(r)
	if err != nil {
		http.Error(w, "invalid start_date or end_date", http.StatusBadRequest)
		return
	}
	period := r.URL.Query().Get("period")
	includeHidden := r.URL.Query().Get("include_hidden") == "true"
	accountIDs, err := parseAccountIDs(r.URL.Query().Get("account_ids"))
	if err != nil {
		http.Error(w, "account_ids must contain valid UUIDs", http.StatusBadRequest)
		return
	}

	report, err := a.analyticsUC.GetSummary(r.Context(), userID, start, end, period, includeHidden, accountIDs)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPeriod) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
// @Param period query string false "Период по умолчанию: day/week/month"
// @Param is_income query boolean false "Тип: доходы(true) или расходы(false)"
// @Param include_hidden query boolean false "Учитывать скрытые транзакции"
// @Param account_ids query string false "CSV список account_id для фильтра"
// @Success 200 {array} domain.CategoryReport
// @Router /api/v1/analytics/categories [get]
func (a *AnalyticsRouter) GetCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	start, end, err := parseDates(r)
	if err != nil {
		http.Error(w, "invalid start_date or end_date", http.StatusBadRequest)
		return
	}
	period := r.URL.Query().Get("period")
	isIncome := r.URL.Query().Get("is_income") == "true"
	includeHidden := r.URL.Query().Get("include_hidden") == "true"
	accountIDs, err := parseAccountIDs(r.URL.Query().Get("account_ids"))
	if err != nil {
		http.Error(w, "account_ids must contain valid UUIDs", http.StatusBadRequest)
		return
	}

	report, err := a.analyticsUC.GetCategoryReport(r.Context(), userID, start, end, period, isIncome, includeHidden, accountIDs)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPeriod) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
// @Param period query string false "Период по умолчанию: day/week/month"
// @Param include_hidden query boolean false "Учитывать скрытые транзакции"
// @Param account_ids query string false "CSV список account_id для фильтра"
// @Success 200 {array} domain.DailyReport
// @Router /api/v1/analytics/daily [get]
func (a *AnalyticsRouter) GetDaily(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isIncome := r.URL.Query().Get("is_income") == "true"
	start, end, err := parseDates(r)
	if err != nil {
		http.Error(w, "invalid start_date or end_date", http.StatusBadRequest)
		return
	}
	period := r.URL.Query().Get("period")
	includeHidden := r.URL.Query().Get("include_hidden") == "true"
	accountIDs, err := parseAccountIDs(r.URL.Query().Get("account_ids"))
	if err != nil {
		http.Error(w, "account_ids must contain valid UUIDs", http.StatusBadRequest)
		return
	}

	report, err := a.analyticsUC.GetDailyDynamics(r.Context(), userID, start, end, period, isIncome, includeHidden, accountIDs)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPeriod) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// @Summary Получить динамику по месяцам
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param is_income query boolean false "Доходы (true) или расходы (false)"
// @Param start_date query string false "Начальная дата (RFC3339)"
// @Param end_date query string false "Конечная дата (RFC3339)"
// @Param period query string false "Период по умолчанию: day/week/month"
// @Param include_hidden query boolean false "Учитывать скрытые транзакции"
// @Param account_ids query string false "CSV список account_id для фильтра"
// @Success 200 {array} domain.MonthlyReport
// @Router /api/v1/analytics/monthly [get]
func (a *AnalyticsRouter) GetMonthly(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isIncome := r.URL.Query().Get("is_income") == "true"
	start, end, err := parseDates(r)
	if err != nil {
		http.Error(w, "invalid start_date or end_date", http.StatusBadRequest)
		return
	}
	period := r.URL.Query().Get("period")
	includeHidden := r.URL.Query().Get("include_hidden") == "true"
	accountIDs, err := parseAccountIDs(r.URL.Query().Get("account_ids"))
	if err != nil {
		http.Error(w, "account_ids must contain valid UUIDs", http.StatusBadRequest)
		return
	}

	report, err := a.analyticsUC.GetMonthlyDynamics(r.Context(), userID, start, end, period, isIncome, includeHidden, accountIDs)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPeriod) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// @Summary Сравнить категории между двумя месяцами
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param first_month query string true "Первый месяц в формате YYYY-MM"
// @Param second_month query string true "Второй месяц в формате YYYY-MM"
// @Param is_income query boolean false "Доходы (true) или расходы (false)"
// @Param include_hidden query boolean false "Учитывать скрытые транзакции"
// @Param account_ids query string false "CSV список account_id для фильтра"
// @Success 200 {array} domain.CategoryCompareReport
// @Router /api/v1/analytics/compare/categories [get]
func (a *AnalyticsRouter) CompareCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	firstMonthRaw := r.URL.Query().Get("first_month")
	secondMonthRaw := r.URL.Query().Get("second_month")
	if firstMonthRaw == "" || secondMonthRaw == "" {
		http.Error(w, "first_month and second_month are required", http.StatusBadRequest)
		return
	}

	firstMonth, err := parseMonth(firstMonthRaw)
	if err != nil {
		http.Error(w, "first_month must be in YYYY-MM format", http.StatusBadRequest)
		return
	}
	secondMonth, err := parseMonth(secondMonthRaw)
	if err != nil {
		http.Error(w, "second_month must be in YYYY-MM format", http.StatusBadRequest)
		return
	}

	isIncome := r.URL.Query().Get("is_income") == "true"
	includeHidden := r.URL.Query().Get("include_hidden") == "true"
	accountIDs, err := parseAccountIDs(r.URL.Query().Get("account_ids"))
	if err != nil {
		http.Error(w, "account_ids must contain valid UUIDs", http.StatusBadRequest)
		return
	}

	report, err := a.analyticsUC.CompareCategoriesByMonths(r.Context(), userID, firstMonth, secondMonth, isIncome, includeHidden, accountIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
