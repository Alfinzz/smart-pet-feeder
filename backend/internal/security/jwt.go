package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"smart-pet-monitoring/backend/internal/domain"
)

type JWTService struct {
	secret    []byte
	expiresIn time.Duration
}

type OwnerClaims struct {
	OwnerID int64  `json:"owner_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, expiresIn time.Duration) *JWTService {
	return &JWTService{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

func (s *JWTService) GenerateOwnerToken(owner domain.Owner) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.expiresIn)
	claims := OwnerClaims{
		OwnerID: owner.ID,
		Email:   owner.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", owner.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *JWTService) VerifyOwnerToken(tokenString string) (*OwnerClaims, error) {
	claims := &OwnerClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrUnauthorized
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrUnauthorized
		}
		return nil, domain.ErrUnauthorized
	}
	if token == nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}
