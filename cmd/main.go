package main

import (
	"Finance-Manager-System/configs"
	"Finance-Manager-System/internal/infrastructure/modules/user/handler"
	"Finance-Manager-System/internal/infrastructure/modules/user/repository"
	"Finance-Manager-System/internal/infrastructure/modules/user/usecase"
	"Finance-Manager-System/internal/infrastructure/postgres"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	cnf := configs.LoadConfig()

	db, err := postgres.NewDB(cnf)
	if err != nil {
		log.Fatal("Problems with db")
	}

	userRepo := repository.NewUserRepository(db)
	useCase := usecase.NewUserCase(userRepo)
	userRouter := handler.NewUserRouter(useCase)

	r := chi.NewRouter()

	r.Mount("/api/v1/users", userRouter.Route())

	serverAddr := net.JoinHostPort(cnf.HttpServer.Adress, cnf.HttpServer.Port)
	log.Printf("server started on: %s\n", serverAddr)

	err = http.ListenAndServe(serverAddr, r)
	if err != nil {
		log.Fatalf("server didnt start: %v", err)
	}
}
