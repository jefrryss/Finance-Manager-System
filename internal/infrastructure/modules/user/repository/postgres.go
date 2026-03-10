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

func (u *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO Users(user_id, email, username, hash_password, created_at, updated_at)
		VALUES(user_id:, email:, username:, hash_password:, created_at:, updated_at:)`

	_, err := u.db.NamedExecContext(ctx, query, user)

	return err
}

func (u *UserRepository) CheckExistUser(ctx context.Context, login string, email string) (bool, error) {
	var exists bool
	params := map[string]interface{}{
		"login": login,
		"email": email,
	}

	query := `SELECT EXISTS(SELECT 1 FROM Users WHERE email = :email AND user_login = :login)`

	query, args, err := u.db.BindNamed(query, params)
	if err != nil {
		return exists, err
	}

	err = u.db.GetContext(ctx, &exists, query, args...)
	return exists, err
}

func (u *UserRepository) UpdateUserInfo(ctx context.Context, user *domain.User) error {
	query := `UPDATE Users SET email = :email, user_login = :login, hash_password = :hash_password, updated_at = :updated_at
	WHERE user_id = :id`

	_, err := u.db.NamedExecContext(ctx, query, user)

	return err

}
