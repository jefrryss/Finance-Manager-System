package router

import (
	"github.com/gin-gonic/gin"

	"expenses-backend/internal/config"
	"expenses-backend/internal/db"
	"expenses-backend/internal/http/handlers"
	"expenses-backend/internal/http/middleware"
)

func Setup(r *gin.Engine, cfg config.Config) {
	r.GET("/health", handlers.Health)

	api := r.Group("/api/v1")

	authH := handlers.AuthHandler{Cfg: cfg, DB: db.DB}
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)

	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg, db.DB))

	protected.POST("/auth/logout", authH.Logout)

	accH := handlers.AccountsHandler{DB: db.DB}
	protected.GET("/accounts", accH.List)
	protected.POST("/accounts", accH.Create)
	protected.GET("/accounts/:id", accH.Get)
	protected.PATCH("/accounts/:id", accH.Patch)
	protected.DELETE("/accounts/:id", accH.Delete)

	catH := handlers.CategoriesHandler{DB: db.DB}
	protected.GET("/categories", catH.List)
	protected.POST("/categories", catH.Create)
	protected.PATCH("/categories/:id", catH.Patch)
	protected.DELETE("/categories/:id", catH.Delete)

	txH := handlers.TransactionsHandler{DB: db.DB}
	protected.GET("/transactions", txH.List)
	protected.POST("/transactions", txH.Create)
	protected.GET("/transactions/:id", txH.Get)
	protected.PATCH("/transactions/:id", txH.Patch)
	protected.DELETE("/transactions/:id", txH.Delete)
}
