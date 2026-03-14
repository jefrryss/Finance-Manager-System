package repository

import (
	"Finance-Manager-System/internal/infrastructure/modules/user/domain"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	query := `INSERT INTO Users(user_id, email, login, hash_password, created_at, updated_at)
		VALUES(:user_id, :email, :login, :hash_password, :created_at, :updated_at)
		RETURNING user_id`

	var id uuid.UUID

	rows, err := u.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return uuid.Nil, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return uuid.Nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (u *UserRepository) CheckExistUser(ctx context.Context, login string, email string) (bool, error) {
	var exists bool
	params := map[string]interface{}{
		"login": login,
		"email": email,
	}

	query := `SELECT EXISTS(SELECT 1 FROM Users WHERE email = :email AND login = :login)`

	query, args, err := u.db.BindNamed(query, params)
	if err != nil {
		return exists, err
	}

	err = u.db.GetContext(ctx, &exists, query, args...)
	return exists, err
}

func (u *UserRepository) UpdateUserInfo(ctx context.Context, id uuid.UUID, hash_password string) error {
	query := `UPDATE Users SET hash_password = :hash_password, updated_at = :updated_at
	WHERE user_id = :id`

	_, err := u.db.NamedExecContext(ctx, query, map[string]interface{}{
		"hash_password": hash_password,
		"id":            id,
		"updated_at":    time.Now(),
	})

	return err

}
