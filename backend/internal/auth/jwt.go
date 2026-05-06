package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/liteoj/liteoj/backend/internal/models"
)

type Claims struct {
	UserID       uint        `json:"uid"`
	Username     string      `json:"usr"`
	Role         models.Role `json:"role"`
	LoginVersion int         `json:"lv"`
	jwt.RegisteredClaims
}

func Issue(secret string, ttl time.Duration, u *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       u.ID,
		Username:     u.Username,
		Role:         u.Role,
		LoginVersion: u.LoginVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

func Parse(secret, tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}
