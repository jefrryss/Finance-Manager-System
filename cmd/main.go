package main

import (
	"log"
	"net"
	"net/http"

	"Finance-Manager-System/configs"
	"Finance-Manager-System/internal/infrastructure/database"
	authMiddleware "Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/postgres"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	// Модуль User
	userHandler "Finance-Manager-System/internal/infrastructure/modules/user/handler"
	userRepo "Finance-Manager-System/internal/infrastructure/modules/user/repository"
	userUC "Finance-Manager-System/internal/infrastructure/modules/user/usecase"

	// Модуль Account
	accountHandler "Finance-Manager-System/internal/infrastructure/modules/account/handler"
	accountRepo "Finance-Manager-System/internal/infrastructure/modules/account/repository"
	accountUC "Finance-Manager-System/internal/infrastructure/modules/account/usecase"

	// Модуль Category
	categoryHandler "Finance-Manager-System/internal/infrastructure/modules/category/handler"
	categoryRepo "Finance-Manager-System/internal/infrastructure/modules/category/repository"
	categoryUC "Finance-Manager-System/internal/infrastructure/modules/category/usecase"

	// Модуль Transaction
	_ "Finance-Manager-System/docs"
	transHandler "Finance-Manager-System/internal/infrastructure/modules/transactions/handler"
	transRepo "Finance-Manager-System/internal/infrastructure/modules/transactions/repository"
	transUC "Finance-Manager-System/internal/infrastructure/modules/transactions/usecase"

	// Модуль Analytics
	analyticsHandler "Finance-Manager-System/internal/infrastructure/modules/analytics/handler"
	analyticsRepo "Finance-Manager-System/internal/infrastructure/modules/analytics/repository"
	analyticsUC "Finance-Manager-System/internal/infrastructure/modules/analytics/usecase"

	// Модуль Recommendations
	recommendationHandler "Finance-Manager-System/internal/infrastructure/modules/recommendations/handler"
	recommendationRepo "Finance-Manager-System/internal/infrastructure/modules/recommendations/repository"
	recommendationUC "Finance-Manager-System/internal/infrastructure/modules/recommendations/usecase"

	// Модуль Goals
	goalHandler "Finance-Manager-System/internal/infrastructure/modules/goals/handler"
	goalRepo "Finance-Manager-System/internal/infrastructure/modules/goals/repository"
	goalUC "Finance-Manager-System/internal/infrastructure/modules/goals/usecase"
)

// @title Finance Manager API
// @version 1.0
// @description API для приложения по управлению личными финансами.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// конфигурации
	cnf := configs.LoadConfig()
	authMiddleware.SetJWTSecret(cnf.JWTSecret)

	// БД
	db, err := postgres.NewDB(cnf)
	if err != nil {
		log.Fatalf("Problems with db: %v", err)
	}

	// менеджер транзакций
	txManager := database.NewTxManager(db)

	// слой Repository
	userRepository := userRepo.NewUserRepository(db)
	accRepository := accountRepo.NewAccountRepo(db)
	catRepository := categoryRepo.NewCategoryRepo(db)
	transactionRepository := transRepo.NewTransRepository(db)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db)
	recommendationsRepository := recommendationRepo.NewRecommendationRepository(db)
	goalsRepository := goalRepo.NewGoalRepo(db)

	// слой UseCase
	userUseCase := userUC.NewUserCase(userRepository, cnf.JWTSecret, catRepository)
	accountUseCase := accountUC.NewAccountUseCase(accRepository)
	transactionUseCase := transUC.NewTransactionUseCase(transactionRepository, accRepository, txManager)
	categoryUseCase := categoryUC.NewCategoryUseCase(catRepository, transactionRepository, txManager)
	analyticsUseCase := analyticsUC.NewAnalyticsUseCase(analyticsRepository)
	recommendationsUseCase := recommendationUC.NewRecommendationUseCase(recommendationsRepository)
	goalsUseCase := goalUC.NewGoalUseCase(goalsRepository, txManager)

	// слой Handler
	userRouter := userHandler.NewUserRouter(userUseCase)
	accountRouter := accountHandler.NewAccountRouter(accountUseCase)
	categoryRouter := categoryHandler.NewCategoryRouter(categoryUseCase)
	transactionRouter := transHandler.NewTransactionRouter(transactionUseCase)
	analyticsRouter := analyticsHandler.NewAnalyticsRouter(analyticsUseCase)
	recommendationRouter := recommendationHandler.NewRecommendationRouter(recommendationsUseCase)
	goalsRouter := goalHandler.NewGoalRouter(goalsUseCase)

	// роутер Chi
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Сборка маршрутов
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/users", userRouter.Route())

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Mount("/accounts", accountRouter.Route())
			r.Mount("/categories", categoryRouter.Route())
			r.Mount("/transactions", transactionRouter.Route())
			r.Mount("/analytics", analyticsRouter.Route())
			r.Mount("/recommendations", recommendationRouter.Route())
			r.Mount("/goals", goalsRouter.Route())
		})
	})

	// Запуск сервера
	serverAddr := net.JoinHostPort(cnf.HttpServer.Adress, cnf.HttpServer.Port)
	log.Printf("Server started on: %s\n", serverAddr)
	log.Printf("Swagger UI is available at: http://%s/swagger/index.html\n", serverAddr)

	err = http.ListenAndServe(serverAddr, r)
	if err != nil {
		log.Fatalf("Server didn't start: %v", err)
	}
}
