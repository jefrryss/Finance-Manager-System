package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists = errors.New("user already exixst")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidLogin      = errors.New("invalid login")
)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type User struct {
	User_id      uuid.UUID `db:"user_id"`
	Email        string    `db:"email"`
	Login        string    `db:"login"`
	HashPassword string    `db:"hash_password"`
	Created_at   time.Time `db:"created_at"`
	Updated_at   time.Time `db:"updated_at"`
}

func NewUser(email string, login string, hashPassword string) (*User, error) {

	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if login == "" {
		return nil, ErrInvalidLogin
	}
	if hashPassword == "" {
		return nil, ErrInvalidPassword
	}
	return &User{
		User_id:      uuid.New(),
		Email:        email,
		Login:        login,
		HashPassword: hashPassword,
		Created_at:   time.Now(),
		Updated_at:   time.Now(),
	}, nil
}
