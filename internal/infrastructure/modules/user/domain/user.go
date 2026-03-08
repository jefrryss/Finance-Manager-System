package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type User struct {
	User_id      uuid.UUID `db:"user_id"`
	Email        string    `db:"email"`
	Login        string    `db:"user_login"`
	HashPassword string    `db:"hash_password"`
	Created_at   time.Time `db:"created_at"`
	Updated_at   time.Time `db:"updated_at"`
}

func NewUser(email string, login string, hashPassword string) (*User, error) {

	if !emailRegex.MatchString(email) {
		return nil, errors.New("email adress uncorrect")
	}
	if login == "" {
		return nil, errors.New("login uncorrect")
	}
	if hashPassword == "" {
		return nil, errors.New("hashPassword uncorrect")
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
