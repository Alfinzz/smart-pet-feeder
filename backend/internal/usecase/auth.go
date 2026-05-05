package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"smart-pet-monitoring/backend/internal/domain"
	"smart-pet-monitoring/backend/internal/security"
)

type OwnerRepository interface {
	Create(ctx context.Context, owner *domain.Owner) error
	FindByEmail(ctx context.Context, email string) (domain.Owner, error)
}

type AuthUsecase struct {
	repo OwnerRepository
	jwt  *security.JWTService
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token     string
	ExpiresAt time.Time
	Owner     domain.Owner
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

func NewAuthUsecase(ownerRepo OwnerRepository, jwtService *security.JWTService) *AuthUsecase {
	return &AuthUsecase{
		repo: ownerRepo,
		jwt:  jwtService,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, input RegisterInput) (LoginOutput, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := strings.TrimSpace(input.Password)
	if name == "" {
		return LoginOutput{}, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if email == "" || password == "" {
		return LoginOutput{}, fmt.Errorf("%w: email and password are required", domain.ErrValidation)
	}
	if len(password) < 8 {
		return LoginOutput{}, fmt.Errorf("%w: password must contain at least 8 characters", domain.ErrValidation)
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return LoginOutput{}, err
	}

	owner := domain.Owner{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	}
	if err := u.repo.Create(ctx, &owner); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return LoginOutput{}, fmt.Errorf("%w: email already registered", domain.ErrConflict)
		}
		return LoginOutput{}, err
	}

	token, expiresAt, err := u.jwt.GenerateOwnerToken(owner)
	if err != nil {
		return LoginOutput{}, err
	}

	owner.PasswordHash = ""
	return LoginOutput{
		Token:     token,
		ExpiresAt: expiresAt,
		Owner:     owner,
	}, nil
}

func (u *AuthUsecase) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" {
		return LoginOutput{}, fmt.Errorf("%w: email and password are required", domain.ErrValidation)
	}

	owner, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return LoginOutput{}, domain.ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if !security.CheckPassword(owner.PasswordHash, password) {
		return LoginOutput{}, domain.ErrInvalidCredentials
	}

	token, expiresAt, err := u.jwt.GenerateOwnerToken(owner)
	if err != nil {
		return LoginOutput{}, err
	}

	owner.PasswordHash = ""
	return LoginOutput{
		Token:     token,
		ExpiresAt: expiresAt,
		Owner:     owner,
	}, nil
}
