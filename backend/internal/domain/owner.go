package domain

import "time"

type Owner struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
