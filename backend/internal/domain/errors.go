package domain

import "errors"

var (
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrValidation         = errors.New("validation error")
)
