package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type CommandRepository struct {
	pool *pgxpool.Pool
}

func (r *CommandRepository) GetNextQueued(ctx context.Context, deviceID string) (domain.ManualCommand, error) {
	const query = `
		UPDATE manual_commands
		SET status = 'sent'
		WHERE id = (
			SELECT id
			FROM manual_commands
			WHERE device_id = $1
			  AND status = 'queued'
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, owner_id, device_id, action, status, created_at
	`

	var command domain.ManualCommand
	err := r.pool.QueryRow(ctx, query, deviceID).Scan(
		&command.ID,
		&command.OwnerID,
		&command.DeviceID,
		&command.Action,
		&command.Status,
		&command.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ManualCommand{}, domain.ErrNotFound
		}
		return domain.ManualCommand{}, err
	}
	return command, nil
}

func (r *CommandRepository) UpdateStatus(ctx context.Context, deviceID string, commandID int64, status domain.CommandStatus) (domain.ManualCommand, error) {
	const query = `
		UPDATE manual_commands
		SET status = $3
		WHERE id = $1
		  AND device_id = $2
		RETURNING id, owner_id, device_id, action, status, created_at
	`

	var command domain.ManualCommand
	err := r.pool.QueryRow(ctx, query, commandID, deviceID, status).Scan(
		&command.ID,
		&command.OwnerID,
		&command.DeviceID,
		&command.Action,
		&command.Status,
		&command.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ManualCommand{}, domain.ErrNotFound
		}
		return domain.ManualCommand{}, err
	}
	return command, nil
}

func NewCommandRepository(pool *pgxpool.Pool) *CommandRepository {
	return &CommandRepository{pool: pool}
}

func (r *CommandRepository) Create(ctx context.Context, command *domain.ManualCommand) error {
	const query = `
		INSERT INTO manual_commands (owner_id, device_id, action, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.pool.QueryRow(ctx, query, command.OwnerID, command.DeviceID, command.Action, command.Status).Scan(
		&command.ID,
		&command.CreatedAt,
	)
}
