package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret []byte
	expiry time.Duration
}

func NewService(secret string, expiry time.Duration) *Service {
	return &Service{secret: []byte(secret), expiry: expiry}
}

type claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func (s *Service) Sign(userID string) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
}

func (s *Service) Verify(tokenStr string) (string, error) {
	var c claims
	t, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !t.Valid {
		return "", errors.New("invalid or expired token")
	}
	return c.UserID, nil
}
