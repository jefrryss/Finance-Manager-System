package main

import (
	"net"
	"net/http"

	"Finance-Manager-System/configs"
	"Finance-Manager-System/internal/infrastructure/cache"
	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/logger"
	authMiddleware "Finance-Manager-System/internal/infrastructure/middleware"
	"Finance-Manager-System/internal/infrastructure/postgres"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

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
	cnf := configs.LoadConfig()
	if err := logger.Init(cnf.Env, cnf.Logger.Dir); err != nil {
		panic(err)
	}
	defer logger.Sync()

	authMiddleware.SetJWTSecret(cnf.JWTSecret)

	db, err := postgres.NewDB(cnf)
	if err != nil {
		zap.L().Fatal("db_connection_failed", zap.Error(err))
	}
	redisCache, redisErr := cache.NewRedisClient(cnf.Redis)
	if redisErr != nil {
		zap.L().Warn("redis_cache_disabled", zap.Error(redisErr))
	} else if redisCache != nil && redisCache.Enabled() {
		zap.L().Info("redis_cache_enabled")
	}

	txManager := database.NewTxManager(db)

	userRepository := userRepo.NewUserRepository(db)
	accRepository := accountRepo.NewAccountRepo(db)
	catRepository := categoryRepo.NewCategoryRepo(db)
	transactionRepository := transRepo.NewTransRepository(db)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db)
	recommendationsRepository := recommendationRepo.NewRecommendationRepository(db)
	goalsRepository := goalRepo.NewGoalRepo(db)

	userUseCase := userUC.NewUserCase(userRepository, cnf.JWTSecret, catRepository)
	accountUseCase := accountUC.NewAccountUseCase(accRepository, catRepository, transactionRepository, txManager)
	transactionUseCase := transUC.NewTransactionUseCase(transactionRepository, accRepository, txManager)
	categoryUseCase := categoryUC.NewCategoryUseCase(catRepository, transactionRepository, txManager)
	analyticsUseCase := analyticsUC.NewAnalyticsUseCase(analyticsRepository)
	recommendationsUseCase := recommendationUC.NewRecommendationUseCase(recommendationsRepository)
	goalsUseCase := goalUC.NewGoalUseCase(goalsRepository, transactionRepository, txManager)

	userRouter := userHandler.NewUserRouter(userUseCase)
	accountRouter := accountHandler.NewAccountRouter(accountUseCase)
	categoryRouter := categoryHandler.NewCategoryRouter(categoryUseCase)
	transactionRouter := transHandler.NewTransactionRouter(transactionUseCase)
	analyticsRouter := analyticsHandler.NewAnalyticsRouter(analyticsUseCase)
	recommendationRouter := recommendationHandler.NewRecommendationRouter(recommendationsUseCase)
	goalsRouter := goalHandler.NewGoalRouter(goalsUseCase)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(authMiddleware.ZapRequestLogger)
	r.Use(authMiddleware.ZapRecoverer)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/users", userRouter.Route())

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Use(authMiddleware.CacheHTTPMiddleware(redisCache))
			r.Mount("/accounts", accountRouter.Route())
			r.Mount("/categories", categoryRouter.Route())
			r.Mount("/transactions", transactionRouter.Route())
			r.Mount("/analytics", analyticsRouter.Route())
			r.Mount("/recommendations", recommendationRouter.Route())
			r.Mount("/goals", goalsRouter.Route())
		})
	})

	serverAddr := net.JoinHostPort(cnf.HttpServer.Adress, cnf.HttpServer.Port)
	zap.L().Info("server_started", zap.String("addr", serverAddr))
	zap.L().Info("swagger_ready", zap.String("url", "http://"+serverAddr+"/swagger/index.html"))

	err = http.ListenAndServe(serverAddr, r)
	if err != nil {
		zap.L().Fatal("server_start_failed", zap.Error(err))
	}
}
