package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type DashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{pool: pool}
}

func (r *DashboardRepository) GetOwner(ctx context.Context, ownerID int64) (domain.OwnerProfile, error) {
	const query = `
		SELECT id, name, email
		FROM owners
		WHERE id = $1
	`

	var owner domain.OwnerProfile
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(
		&owner.ID,
		&owner.Name,
		&owner.Email,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OwnerProfile{}, domain.ErrNotFound
		}
		return domain.OwnerProfile{}, err
	}
	return owner, nil
}

func (r *DashboardRepository) GetPetProfile(ctx context.Context, ownerID int64) (domain.PetProfile, error) {
	const query = `
		SELECT
			id,
			owner_id,
			device_id,
			name,
			species,
			breed,
			age_years,
			weight_kg,
			daily_feed_target_grams,
			health_score,
			health_status,
			health_headline,
			health_description,
			activity_minutes,
			sleep_hours,
			COALESCE(photo_path, '')
		FROM pets
		WHERE owner_id = $1
		ORDER BY id
		LIMIT 1
	`

	var pet domain.PetProfile
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(
		&pet.ID,
		&pet.OwnerID,
		&pet.DeviceID,
		&pet.Name,
		&pet.Species,
		&pet.Breed,
		&pet.AgeYears,
		&pet.WeightKG,
		&pet.DailyFeedTargetGrams,
		&pet.HealthScore,
		&pet.HealthStatus,
		&pet.HealthHeadline,
		&pet.HealthDescription,
		&pet.ActivityMinutes,
		&pet.SleepHours,
		&pet.PhotoPath,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PetProfile{}, domain.ErrNotFound
		}
		return domain.PetProfile{}, err
	}
	return pet, nil
}

func (r *DashboardRepository) GetDeviceStatus(ctx context.Context, deviceID string) (domain.DeviceStatus, error) {
	const query = `
		SELECT id, name, food_stock_percent, water_available, water_status, last_seen_at
		FROM devices
		WHERE id = $1
	`

	var status domain.DeviceStatus
	err := r.pool.QueryRow(ctx, query, deviceID).Scan(
		&status.ID,
		&status.Name,
		&status.FoodStockPercent,
		&status.WaterAvailable,
		&status.WaterStatus,
		&status.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DeviceStatus{}, domain.ErrNotFound
		}
		return domain.DeviceStatus{}, err
	}
	return status, nil
}

func (r *DashboardRepository) GetDailyConsumption(ctx context.Context, deviceID string, days int) ([]domain.DailyConsumption, error) {
	const query = `
		WITH days AS (
			SELECT generate_series(
				CURRENT_DATE - (($2::int - 1) * INTERVAL '1 day'),
				CURRENT_DATE,
				INTERVAL '1 day'
			)::date AS day
		)
		SELECT
			days.day::timestamptz,
			COALESCE(SUM(feed_readings.weight_grams), 0)
		FROM days
		LEFT JOIN feed_readings
			ON feed_readings.device_id = $1
			AND feed_readings.recorded_at >= days.day::timestamptz
			AND feed_readings.recorded_at < (days.day + 1)::timestamptz
		GROUP BY days.day
		ORDER BY days.day
	`

	rows, err := r.pool.Query(ctx, query, deviceID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.DailyConsumption, 0, days)
	for rows.Next() {
		var item domain.DailyConsumption
		if err := rows.Scan(&item.Date, &item.TotalGrams); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DashboardRepository) ListUpcomingTasks(ctx context.Context, petID int64, limit int) ([]domain.CareTask, error) {
	const query = `
		SELECT id, pet_id, category, title, subtitle, due_label, due_at, priority, sort_order
		FROM care_tasks
		WHERE pet_id = $1
		ORDER BY sort_order ASC, due_at ASC NULLS LAST, id ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, petID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.CareTask, 0, limit)
	for rows.Next() {
		task, err := scanCareTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, domain.ErrNotFound
	}
	return tasks, nil
}

func (r *DashboardRepository) CreateCareTask(ctx context.Context, petID int64, input domain.CareTaskInput) (domain.CareTask, error) {
	const query = `
		INSERT INTO care_tasks (pet_id, category, title, subtitle, due_label, due_at, priority, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, pet_id, category, title, subtitle, due_label, due_at, priority, sort_order
	`

	task, err := scanCareTaskRow(r.pool.QueryRow(
		ctx,
		query,
		petID,
		input.Category,
		input.Title,
		input.Subtitle,
		input.DueLabel,
		input.DueAt,
		input.Priority,
		input.SortOrder,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.CareTask{}, domain.ErrConflict
		}
		return domain.CareTask{}, err
	}
	return task, nil
}

func (r *DashboardRepository) UpdateCareTask(ctx context.Context, ownerID, taskID int64, input domain.CareTaskInput) (domain.CareTask, error) {
	const query = `
		UPDATE care_tasks
		SET
			category = $3,
			title = $4,
			subtitle = $5,
			due_label = $6,
			due_at = $7,
			priority = $8,
			sort_order = $9
		WHERE id = $1
		  AND EXISTS (
			SELECT 1
			FROM pets
			WHERE pets.id = care_tasks.pet_id
			  AND pets.owner_id = $2
		  )
		RETURNING id, pet_id, category, title, subtitle, due_label, due_at, priority, sort_order
	`

	task, err := scanCareTaskRow(r.pool.QueryRow(
		ctx,
		query,
		taskID,
		ownerID,
		input.Category,
		input.Title,
		input.Subtitle,
		input.DueLabel,
		input.DueAt,
		input.Priority,
		input.SortOrder,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CareTask{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.CareTask{}, domain.ErrConflict
		}
		return domain.CareTask{}, err
	}
	return task, nil
}

func (r *DashboardRepository) DeleteCareTask(ctx context.Context, ownerID, taskID int64) error {
	const query = `
		DELETE FROM care_tasks
		WHERE id = $1
		  AND EXISTS (
			SELECT 1
			FROM pets
			WHERE pets.id = care_tasks.pet_id
			  AND pets.owner_id = $2
		  )
	`

	tag, err := r.pool.Exec(ctx, query, taskID, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DashboardRepository) UpdateDeviceSettings(
	ctx context.Context,
	deviceID string,
	name string,
	foodStockPercent *float64,
	waterAvailable *bool,
	waterStatus string,
) (domain.DeviceStatus, error) {
	const query = `
		UPDATE devices
		SET
			name = COALESCE(NULLIF($2, ''), name),
			food_stock_percent = COALESCE($3::double precision, food_stock_percent),
			water_available = COALESCE($4::boolean, water_available),
			water_status = COALESCE(NULLIF($5, ''), water_status)
		WHERE id = $1
		RETURNING id, name, food_stock_percent, water_available, water_status, last_seen_at
	`

	var status domain.DeviceStatus
	err := r.pool.QueryRow(ctx, query, deviceID, name, foodStockPercent, waterAvailable, waterStatus).Scan(
		&status.ID,
		&status.Name,
		&status.FoodStockPercent,
		&status.WaterAvailable,
		&status.WaterStatus,
		&status.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DeviceStatus{}, domain.ErrNotFound
		}
		return domain.DeviceStatus{}, err
	}
	return status, nil
}

func (r *DashboardRepository) UpsertPetProfile(ctx context.Context, ownerID int64, input domain.PetProfileUpdate) (domain.PetProfile, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PetProfile{}, err
	}
	defer tx.Rollback(ctx)

	const deviceQuery = `
		INSERT INTO devices (id)
		VALUES ($1)
		ON CONFLICT (id) DO NOTHING
	`
	if _, err := tx.Exec(ctx, deviceQuery, input.DeviceID); err != nil {
		return domain.PetProfile{}, err
	}

	const petQuery = `
		INSERT INTO pets (
			owner_id,
			device_id,
			name,
			species,
			breed,
			age_years,
			weight_kg,
			daily_feed_target_grams,
			health_score,
			health_status,
			health_headline,
			health_description,
			activity_minutes,
			sleep_hours
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (owner_id) DO UPDATE SET
			device_id = EXCLUDED.device_id,
			name = EXCLUDED.name,
			species = EXCLUDED.species,
			breed = EXCLUDED.breed,
			age_years = EXCLUDED.age_years,
			weight_kg = EXCLUDED.weight_kg,
			daily_feed_target_grams = EXCLUDED.daily_feed_target_grams,
			health_score = EXCLUDED.health_score,
			health_status = EXCLUDED.health_status,
			health_headline = EXCLUDED.health_headline,
			health_description = EXCLUDED.health_description,
			activity_minutes = EXCLUDED.activity_minutes,
			sleep_hours = EXCLUDED.sleep_hours
		RETURNING
			id,
			owner_id,
			device_id,
			name,
			species,
			breed,
			age_years,
			weight_kg,
			daily_feed_target_grams,
			health_score,
			health_status,
			health_headline,
			health_description,
			activity_minutes,
			sleep_hours,
			COALESCE(photo_path, '')
	`

	var pet domain.PetProfile
	if err := tx.QueryRow(
		ctx,
		petQuery,
		ownerID,
		input.DeviceID,
		input.Name,
		input.Species,
		input.Breed,
		input.AgeYears,
		input.WeightKG,
		input.DailyFeedTargetGrams,
		input.HealthScore,
		input.HealthStatus,
		input.HealthHeadline,
		input.HealthDescription,
		input.ActivityMinutes,
		input.SleepHours,
	).Scan(
		&pet.ID,
		&pet.OwnerID,
		&pet.DeviceID,
		&pet.Name,
		&pet.Species,
		&pet.Breed,
		&pet.AgeYears,
		&pet.WeightKG,
		&pet.DailyFeedTargetGrams,
		&pet.HealthScore,
		&pet.HealthStatus,
		&pet.HealthHeadline,
		&pet.HealthDescription,
		&pet.ActivityMinutes,
		&pet.SleepHours,
		&pet.PhotoPath,
	); err != nil {
		return domain.PetProfile{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.PetProfile{}, err
	}
	return pet, nil
}

func (r *DashboardRepository) UpdatePetPhoto(ctx context.Context, ownerID int64, photoPath string) (domain.PetProfile, string, error) {
	const query = `
		WITH previous AS (
			SELECT COALESCE(photo_path, '') AS old_photo_path
			FROM pets
			WHERE owner_id = $1
		),
		updated AS (
			UPDATE pets
			SET photo_path = $2
			WHERE owner_id = $1
			RETURNING
				id,
				owner_id,
				device_id,
				name,
				species,
				breed,
				age_years,
				weight_kg,
				daily_feed_target_grams,
				health_score,
				health_status,
				health_headline,
				health_description,
				activity_minutes,
				sleep_hours,
				COALESCE(photo_path, '') AS photo_path
		)
		SELECT
			updated.id,
			updated.owner_id,
			updated.device_id,
			updated.name,
			updated.species,
			updated.breed,
			updated.age_years,
			updated.weight_kg,
			updated.daily_feed_target_grams,
			updated.health_score,
			updated.health_status,
			updated.health_headline,
			updated.health_description,
			updated.activity_minutes,
			updated.sleep_hours,
			updated.photo_path,
			previous.old_photo_path
		FROM updated
		JOIN previous ON TRUE
	`

	var pet domain.PetProfile
	var oldPhotoPath string
	err := r.pool.QueryRow(ctx, query, ownerID, photoPath).Scan(
		&pet.ID,
		&pet.OwnerID,
		&pet.DeviceID,
		&pet.Name,
		&pet.Species,
		&pet.Breed,
		&pet.AgeYears,
		&pet.WeightKG,
		&pet.DailyFeedTargetGrams,
		&pet.HealthScore,
		&pet.HealthStatus,
		&pet.HealthHeadline,
		&pet.HealthDescription,
		&pet.ActivityMinutes,
		&pet.SleepHours,
		&pet.PhotoPath,
		&oldPhotoPath,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PetProfile{}, "", domain.ErrNotFound
		}
		return domain.PetProfile{}, "", err
	}
	return pet, oldPhotoPath, nil
}

type careTaskScanner interface {
	Scan(dest ...any) error
}

func scanCareTaskRow(row careTaskScanner) (domain.CareTask, error) {
	var task domain.CareTask
	var dueAt pgtype.Date
	if err := row.Scan(
		&task.ID,
		&task.PetID,
		&task.Category,
		&task.Title,
		&task.Subtitle,
		&task.DueLabel,
		&dueAt,
		&task.Priority,
		&task.SortOrder,
	); err != nil {
		return domain.CareTask{}, err
	}
	if dueAt.Valid {
		value := dueAt.Time
		task.DueAt = &value
	}
	return task, nil
}

func scanCareTask(rows pgx.Rows) (domain.CareTask, error) {
	return scanCareTaskRow(rows)
}
