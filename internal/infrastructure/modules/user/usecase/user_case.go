package usecase

import (
	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"Finance-Manager-System/internal/infrastructure/modules/user/repository"
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserCase struct {
	db *repository.UserRepository
}

func NewUserCase(db *repository.UserRepository) *UserCase {
	return &UserCase{
		db: db,
	}
}

func (u *UserCase) RegistrateUser(ctx context.Context, email string, login string, password string) error {

	bytesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	login = strings.ToLower(strings.TrimSpace(login))

	exists_user, err := u.db.CheckExistUser(ctx, login, email)
	if err != nil {
		return err
	}
	if exists_user {
		return fmt.Errorf("User already exists with login: %s and email: %s", login, email)
	}

	user, err := domain.NewUser(email, login, string(bytesPassword))
	if err != nil {
		return err
	}

	err = u.db.CreateUser(ctx, user)
	return err
}

func (u *UserCase) ChangeUserPassword(ctx context.Context, user *domain.User, password string) error {
	bytesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.HashPassword = string(bytesPassword)

	//Добавлять или нет проверку на то что пользователь сущ?

	err = u.db.UpdateUserInfo(ctx, user)
	return err
}
