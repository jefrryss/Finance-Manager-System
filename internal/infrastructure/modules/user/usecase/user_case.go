package usecase

import (
	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"Finance-Manager-System/internal/infrastructure/modules/user/repository"
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserCase struct {
	db           *repository.UserRepository
	jwtSecretKey []byte
	catBootstrap DefaultCategoryBootstrapper
}

type DefaultCategoryBootstrapper interface {
	EnsureDefaultCategories(ctx context.Context, userID uuid.UUID) error
}

func NewUserCase(db *repository.UserRepository, jwtSecret string, catBootstrap DefaultCategoryBootstrapper) *UserCase {
	return &UserCase{
		db:           db,
		jwtSecretKey: []byte(jwtSecret),
		catBootstrap: catBootstrap,
	}
}

func (u *UserCase) LoginUser(ctx context.Context, identifier, password string) (string, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))

	user, err := u.db.GetUserByEmailOrLogin(ctx, identifier)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	if err := u.catBootstrap.EnsureDefaultCategories(ctx, user.User_id); err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.User_id.String(),
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(u.jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.catBootstrap.EnsureDefaultCategories(ctx, id); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *UserCase) ChangeUserPassword(ctx context.Context, id uuid.UUID, password string) error {
	bytesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = u.db.UpdateUserInfo(ctx, id, string(bytesPassword))
	return err
}
