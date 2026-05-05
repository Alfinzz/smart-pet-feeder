package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type FeedRepository struct {
	pool *pgxpool.Pool
}

func NewFeedRepository(pool *pgxpool.Pool) *FeedRepository {
	return &FeedRepository{pool: pool}
}

func (r *FeedRepository) Create(ctx context.Context, reading *domain.FeedReading) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insertReadingQuery = `
		INSERT INTO feed_readings (device_id, weight_grams, recorded_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	if err := tx.QueryRow(ctx, insertReadingQuery, reading.DeviceID, reading.WeightGrams, reading.RecordedAt).Scan(
		&reading.ID,
		&reading.CreatedAt,
	); err != nil {
		return err
	}

	const upsertDeviceQuery = `
		INSERT INTO devices (id, last_seen_at, food_stock_percent, water_available, water_status)
		VALUES (
			$1,
			$2,
			COALESCE($3::double precision, 75),
			COALESCE($4::boolean, TRUE),
			COALESCE(NULLIF($5, ''), 'Clean & Fresh')
		)
		ON CONFLICT (id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			food_stock_percent = COALESCE($3::double precision, devices.food_stock_percent),
			water_available = COALESCE($4::boolean, devices.water_available),
			water_status = COALESCE(NULLIF($5, ''), devices.water_status)
	`

	if _, err := tx.Exec(
		ctx,
		upsertDeviceQuery,
		reading.DeviceID,
		reading.RecordedAt,
		reading.FoodStockPercent,
		reading.WaterAvailable,
		reading.WaterStatus,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *FeedRepository) ListHistory(ctx context.Context, filter domain.FeedHistoryFilter) ([]domain.FeedReading, error) {
	const query = `
		SELECT id, device_id, weight_grams, recorded_at, created_at
		FROM feed_readings
		WHERE ($1::timestamptz IS NULL OR recorded_at >= $1::timestamptz)
		  AND ($2::timestamptz IS NULL OR recorded_at <= $2::timestamptz)
		ORDER BY recorded_at DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, filter.From, filter.To, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := make([]domain.FeedReading, 0)
	for rows.Next() {
		var reading domain.FeedReading
		if err := rows.Scan(
			&reading.ID,
			&reading.DeviceID,
			&reading.WeightGrams,
			&reading.RecordedAt,
			&reading.CreatedAt,
		); err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}

func (r *FeedRepository) UpsertDeviceStatus(ctx context.Context, update domain.DeviceStatusUpdate) (domain.DeviceStatus, error) {
	const query = `
		INSERT INTO devices (id, last_seen_at, food_stock_percent, water_available, water_status)
		VALUES (
			$1,
			$2,
			COALESCE($3::double precision, 75),
			COALESCE($4::boolean, TRUE),
			COALESCE(NULLIF($5, ''), 'Clean & Fresh')
		)
		ON CONFLICT (id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			food_stock_percent = COALESCE($3::double precision, devices.food_stock_percent),
			water_available = COALESCE($4::boolean, devices.water_available),
			water_status = COALESCE(NULLIF($5, ''), devices.water_status)
		RETURNING id, name, food_stock_percent, water_available, water_status, last_seen_at
	`

	var status domain.DeviceStatus
	err := r.pool.QueryRow(
		ctx,
		query,
		update.ID,
		update.LastSeenAt,
		update.FoodStockPercent,
		update.WaterAvailable,
		update.WaterStatus,
	).Scan(
		&status.ID,
		&status.Name,
		&status.FoodStockPercent,
		&status.WaterAvailable,
		&status.WaterStatus,
		&status.LastSeenAt,
	)
	if err != nil {
		return domain.DeviceStatus{}, err
	}
	return status, nil
}
