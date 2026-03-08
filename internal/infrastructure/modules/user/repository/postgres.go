package repository

import (
	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"context"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) CreateUser(ctx context.Context, user domain.User) error {
	query := `INSERT INTO Users (user_id, email, username, hash_password, 
				created_at, updated_at)`

	_, err := u.db.NamedExecContext(ctx, query, user)

	return err
}

func (u *UserRepository) SearchByLogin(ctx context.Context, login string) error {

	params := map[string]interface{}{
		"login": login,
	}

	query := `SELECT EXISTS(SELECT 1 FROM Users WHERE user_login = :login)`

	_, err := u.db.NamedExecContext(ctx, query, params)

	return err
}
