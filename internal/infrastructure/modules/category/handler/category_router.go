package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/modules/category/usecase"
)

type CategoryRouter struct {
	categoryUC *usecase.CategoryUseCase
}

func NewCategoryRouter(categoryUC *usecase.CategoryUseCase) *CategoryRouter {
	return &CategoryRouter{
		categoryUC: categoryUC,
	}
}

func (c *CategoryRouter) Route() chi.Router {
	r := chi.NewRouter()

	r.Post("/", c.CreateCategory)
	r.Get("/", c.GetCategories)
	r.Put("/{id}", c.UpdateCategory)
	r.Delete("/{id}", c.DeleteCategory)

	return r
}

type CreateCategoryReq struct {
	Name     string  `json:"name"`
	IsIncome bool    `json:"is_income"`
	IconURL  *string `json:"icon_url"`
}

type UpdateCategoryReq struct {
	Name    string  `json:"name"`
	IconURL *string `json:"icon_url"`
}

type DeleteCategoryReq struct {
	ReplacementCategoryID *uuid.UUID `json:"replacement_category_id"`
}

// @Summary Создать категорию
// @Tags categories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body CreateCategoryReq true "Данные категории"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/categories [post]
func (c *CategoryRouter) CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateCategoryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	id, err := c.categoryUC.CreateCustomCategory(r.Context(), userID, req.Name, req.IsIncome, req.IconURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"category_id": id,
	})
}

// @Summary Получить категории
// @Tags categories
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} domain.Category
// @Router /api/v1/categories [get]
func (c *CategoryRouter) GetCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	categories, err := c.categoryUC.GetUserCategories(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// @Summary Обновить категорию
// @Tags categories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID категории"
// @Param request body UpdateCategoryReq true "Новые данные"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/categories/{id} [put]
func (c *CategoryRouter) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	catID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req UpdateCategoryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = c.categoryUC.UpdateCategory(r.Context(), userID, catID, req.Name, req.IconURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

// @Summary Удалить категорию
// @Tags categories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID категории"
// @Param request body DeleteCategoryReq false "Опционально ID для переноса"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/categories/{id} [delete]
func (c *CategoryRouter) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	catID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req DeleteCategoryReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	err = c.categoryUC.DeleteCategory(r.Context(), userID, catID, req.ReplacementCategoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
