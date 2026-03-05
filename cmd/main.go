package main

import (
	"Finance-Manager-System/configs"
	"Finance-Manager-System/internal/infrastructure/modules/user/repository"
	"Finance-Manager-System/internal/infrastructure/modules/user/usecase"
	"Finance-Manager-System/internal/infrastructure/postgres"
	"context"
	"fmt"
)

func main() {
	cnf := configs.LoadConfig()
	db, err := postgres.NewDB(cnf)
	if err != nil {
		fmt.Println(err)
	}
	userRep := repository.NewUserRepository(db)
	useCase := usecase.NewUserCase(userRep)
	ctx := context.Background()
	useCase.RegistrateUser(ctx, "lofin@gmail.com", "User1234", "123123")
}
