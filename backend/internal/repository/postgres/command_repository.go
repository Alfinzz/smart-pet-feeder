package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type CommandRepository struct {
	pool *pgxpool.Pool
}

func (r *CommandRepository) GetByOwner(ctx context.Context, ownerID, commandID int64) (domain.ManualCommand, error) {
	const query = `
		SELECT id, owner_id, device_id, action, status, attempt_count, last_error, created_at, updated_at, completed_at
		FROM manual_commands
		WHERE id = $1
		  AND owner_id = $2
	`

	command, err := scanManualCommand(r.pool.QueryRow(ctx, query, commandID, ownerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ManualCommand{}, domain.ErrNotFound
		}
		return domain.ManualCommand{}, err
	}
	return command, nil
}

func (r *CommandRepository) GetNextQueued(ctx context.Context, deviceID string, staleBefore time.Time, maxAttempts int) (domain.ManualCommand, error) {
	const query = `
		WITH failed_stale AS (
			UPDATE manual_commands
			SET
				status = 'failed',
				last_error = 'device did not confirm before retry limit',
				updated_at = NOW()
			WHERE device_id = $1
			  AND status = 'sent'
			  AND updated_at <= $2
			  AND attempt_count >= $3
		),
		active_sent AS (
			SELECT id
			FROM manual_commands
			WHERE device_id = $1
			  AND status = 'sent'
			  AND updated_at > $2
			LIMIT 1
		),
		next_command AS (
			SELECT id
			FROM manual_commands
			WHERE device_id = $1
			  AND NOT EXISTS (SELECT 1 FROM active_sent)
			  AND (
				status = 'queued'
				OR (status = 'sent' AND updated_at <= $2 AND attempt_count < $3)
			  )
			ORDER BY
				CASE WHEN status = 'sent' THEN 0 ELSE 1 END,
				created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE manual_commands
		SET
			status = 'sent',
			attempt_count = attempt_count + 1,
			last_error = '',
			updated_at = NOW()
		WHERE id = (SELECT id FROM next_command)
		RETURNING id, owner_id, device_id, action, status, attempt_count, last_error, created_at, updated_at, completed_at
	`

	command, err := scanManualCommand(r.pool.QueryRow(ctx, query, deviceID, staleBefore, maxAttempts))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ManualCommand{}, domain.ErrNotFound
		}
		return domain.ManualCommand{}, err
	}
	return command, nil
}

func (r *CommandRepository) UpdateStatus(ctx context.Context, deviceID string, commandID int64, status domain.CommandStatus, lastError string) (domain.ManualCommand, error) {
	const query = `
		UPDATE manual_commands
		SET
			status = $3,
			last_error = CASE WHEN $3 = 'failed' THEN $4 ELSE '' END,
			updated_at = NOW(),
			completed_at = CASE WHEN $3 = 'completed' THEN NOW() ELSE completed_at END
		WHERE id = $1
		  AND device_id = $2
		  AND status = 'sent'
		RETURNING id, owner_id, device_id, action, status, attempt_count, last_error, created_at, updated_at, completed_at
	`

	command, err := scanManualCommand(r.pool.QueryRow(ctx, query, commandID, deviceID, status, lastError))
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
		RETURNING id, owner_id, device_id, action, status, attempt_count, last_error, created_at, updated_at, completed_at
	`

	created, err := scanManualCommand(r.pool.QueryRow(ctx, query, command.OwnerID, command.DeviceID, command.Action, command.Status))
	if err != nil {
		return err
	}
	*command = created
	return nil
}

func (r *CommandRepository) GetOwnerDeviceID(ctx context.Context, ownerID int64) (string, error) {
	const query = `
		SELECT device_id
		FROM pets
		WHERE owner_id = $1
		ORDER BY id
		LIMIT 1
	`
	var deviceID string
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(&deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	return deviceID, nil
}

func (r *CommandRepository) OwnerHasDevice(ctx context.Context, ownerID int64, deviceID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM pets
			WHERE owner_id = $1
			  AND device_id = $2
		)
	`
	var ok bool
	if err := r.pool.QueryRow(ctx, query, ownerID, deviceID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func scanManualCommand(row pgx.Row) (domain.ManualCommand, error) {
	var command domain.ManualCommand
	var completedAt pgtype.Timestamptz
	err := row.Scan(
		&command.ID,
		&command.OwnerID,
		&command.DeviceID,
		&command.Action,
		&command.Status,
		&command.AttemptCount,
		&command.LastError,
		&command.CreatedAt,
		&command.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		return domain.ManualCommand{}, err
	}
	if completedAt.Valid {
		command.CompletedAt = &completedAt.Time
	}
	return command, nil
}
