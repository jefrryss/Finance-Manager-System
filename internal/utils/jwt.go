package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	JTI string `json:"jti"`
}

func NewToken(userID uuid.UUID, secret string, ttl time.Duration) (tokenString string, jti string, exp time.Time, err error) {
	jti = uuid.NewString()
	exp = time.Now().UTC().Add(ttl)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		JTI: jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(secret))
	return
}

func ParseToken(tokenString string, secret string) (userID uuid.UUID, jti string, exp time.Time, err error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return uuid.Nil, "", time.Time{}, err
	}

	cl, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return uuid.Nil, "", time.Time{}, errors.New("invalid token")
	}

	uid, err := uuid.Parse(cl.Subject)
	if err != nil {
		return uuid.Nil, "", time.Time{}, err
	}

	exp = time.Time{}
	if cl.ExpiresAt != nil {
		exp = cl.ExpiresAt.Time
	}
	return uid, cl.JTI, exp, nil
}
