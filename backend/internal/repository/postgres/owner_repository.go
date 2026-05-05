package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type OwnerRepository struct {
	pool *pgxpool.Pool
}

func NewOwnerRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool}
}

func (r *OwnerRepository) FindByEmail(ctx context.Context, email string) (domain.Owner, error) {
	const query = `
		SELECT id, name, email, password_hash, created_at
		FROM owners
		WHERE email = $1
	`

	var owner domain.Owner
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&owner.ID,
		&owner.Name,
		&owner.Email,
		&owner.PasswordHash,
		&owner.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Owner{}, domain.ErrNotFound
		}
		return domain.Owner{}, err
	}
	return owner, nil
}

func (r *OwnerRepository) Create(ctx context.Context, owner *domain.Owner) error {
	const query = `
		INSERT INTO owners (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query, owner.Name, owner.Email, owner.PasswordHash).Scan(
		&owner.ID,
		&owner.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}
