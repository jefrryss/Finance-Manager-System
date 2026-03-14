package usecase

import (
	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"Finance-Manager-System/internal/infrastructure/modules/user/repository"
	"context"
	"strings"

	"github.com/google/uuid"
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

func (u *UserCase) RegistrateUser(ctx context.Context, email string, login string, password string) (uuid.UUID, error) {

	bytesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	login = strings.ToLower(strings.TrimSpace(login))

	exists_user, err := u.db.CheckExistUser(ctx, login, email)
	if err != nil {
		return uuid.Nil, err
	}
	if exists_user {
		return uuid.Nil, domain.ErrUserAlreadyExists
	}

	user, err := domain.NewUser(email, login, string(bytesPassword))
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.db.CreateUser(ctx, user)
	return id, err
}

func (u *UserCase) ChangeUserPassword(ctx context.Context, id uuid.UUID, password string) error {
	bytesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = u.db.UpdateUserInfo(ctx, id, string(bytesPassword))
	return err
}
