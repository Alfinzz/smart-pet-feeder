package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type HealthRepository struct {
	pool *pgxpool.Pool
}

func NewHealthRepository(pool *pgxpool.Pool) *HealthRepository {
	return &HealthRepository{pool: pool}
}

func (r *HealthRepository) GetPrimaryHealthPet(ctx context.Context, ownerID int64) (domain.HealthPet, error) {
	const query = `
		SELECT id, owner_id, device_id, name, weight_kg, activity_minutes, sleep_hours, daily_feed_target_grams
		FROM pets
		WHERE owner_id = $1
		ORDER BY id
		LIMIT 1
	`

	var pet domain.HealthPet
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(
		&pet.ID,
		&pet.OwnerID,
		&pet.DeviceID,
		&pet.Name,
		&pet.WeightKG,
		&pet.ActivityMinutes,
		&pet.SleepHours,
		&pet.DailyFeedTargetGrams,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.HealthPet{}, domain.ErrNotFound
		}
		return domain.HealthPet{}, err
	}
	return pet, nil
}

func (r *HealthRepository) SaveVitalSigns(ctx context.Context, petID int64, input domain.VitalSignsInput) (domain.VitalSigns, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.VitalSigns{}, err
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO pet_vital_signs (pet_id, weight_kg, activity_minutes, sleep_hours, recorded_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, pet_id, weight_kg, activity_minutes, sleep_hours, recorded_at, created_at
	`

	var vitals domain.VitalSigns
	err = tx.QueryRow(
		ctx,
		query,
		petID,
		input.WeightKG,
		input.ActivityMinutes,
		input.SleepHours,
		input.RecordedAt,
	).Scan(
		&vitals.ID,
		&vitals.PetID,
		&vitals.WeightKG,
		&vitals.ActivityMinutes,
		&vitals.SleepHours,
		&vitals.RecordedAt,
		&vitals.CreatedAt,
	)
	if err != nil {
		return domain.VitalSigns{}, err
	}

	const updatePetQuery = `
		UPDATE pets
		SET weight_kg = $2,
		    activity_minutes = $3,
		    sleep_hours = $4
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updatePetQuery, petID, input.WeightKG, input.ActivityMinutes, input.SleepHours); err != nil {
		return domain.VitalSigns{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.VitalSigns{}, err
	}
	return vitals, nil
}

func (r *HealthRepository) GetLatestVitalSigns(ctx context.Context, petID int64) (domain.VitalSigns, error) {
	const query = `
		SELECT id, pet_id, weight_kg, activity_minutes, sleep_hours, recorded_at, created_at
		FROM pet_vital_signs
		WHERE pet_id = $1
		ORDER BY recorded_at DESC, id DESC
		LIMIT 1
	`
	return r.scanVitalSigns(ctx, query, petID)
}

func (r *HealthRepository) GetPreviousVitalSigns(ctx context.Context, petID, latestID int64) (domain.VitalSigns, error) {
	const query = `
		SELECT id, pet_id, weight_kg, activity_minutes, sleep_hours, recorded_at, created_at
		FROM pet_vital_signs
		WHERE pet_id = $1
		  AND id <> $2
		ORDER BY recorded_at DESC, id DESC
		LIMIT 1
	`
	return r.scanVitalSigns(ctx, query, petID, latestID)
}

func (r *HealthRepository) AverageDailyFeedConsumption(ctx context.Context, deviceID string, days int) (float64, error) {
	const query = `
		WITH days AS (
			SELECT generate_series(
				CURRENT_DATE - (($2::int - 1) * INTERVAL '1 day'),
				CURRENT_DATE,
				INTERVAL '1 day'
			)::date AS day
		),
		daily_totals AS (
			SELECT
				days.day,
				COALESCE(SUM(feed_readings.weight_grams), 0) AS total_grams
			FROM days
			LEFT JOIN feed_readings
				ON feed_readings.device_id = $1
				AND feed_readings.recorded_at >= days.day::timestamptz
				AND feed_readings.recorded_at < (days.day + 1)::timestamptz
			GROUP BY days.day
		)
		SELECT COALESCE(AVG(total_grams), 0)
		FROM daily_totals
	`

	var average float64
	if err := r.pool.QueryRow(ctx, query, deviceID, days).Scan(&average); err != nil {
		return 0, err
	}
	return average, nil
}

func (r *HealthRepository) CountOverduePendingTasks(ctx context.Context, petID int64) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM pet_tasks
		WHERE pet_id = $1
		  AND status = 'pending'
		  AND due_date < CURRENT_DATE
	`

	var count int
	if err := r.pool.QueryRow(ctx, query, petID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *HealthRepository) scanVitalSigns(ctx context.Context, query string, args ...any) (domain.VitalSigns, error) {
	var vitals domain.VitalSigns
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&vitals.ID,
		&vitals.PetID,
		&vitals.WeightKG,
		&vitals.ActivityMinutes,
		&vitals.SleepHours,
		&vitals.RecordedAt,
		&vitals.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VitalSigns{}, domain.ErrNotFound
		}
		return domain.VitalSigns{}, err
	}
	return vitals, nil
}
