package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) GetPetDetails(ctx context.Context, ownerID int64) (domain.PetDetails, error) {
	const query = `
		SELECT id, owner_id, device_id, name, species, breed, gender, birth_date, daily_feed_target_grams, created_at
		FROM pets
		WHERE owner_id = $1
		ORDER BY id
		LIMIT 1
	`
	return r.scanPetDetails(ctx, query, ownerID)
}

func (r *ProfileRepository) CreatePetDetails(ctx context.Context, ownerID int64, input domain.PetDetailsInput) (domain.PetDetails, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PetDetails{}, err
	}
	defer tx.Rollback(ctx)

	if err := ensureDevice(ctx, tx, input.DeviceID); err != nil {
		return domain.PetDetails{}, err
	}

	const query = `
		INSERT INTO pets (
			owner_id,
			device_id,
			name,
			species,
			breed,
			gender,
			birth_date,
			daily_feed_target_grams,
			health_score,
			health_status,
			health_headline,
			health_description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 80, 'Good', 'Health data ready', 'Update vital signs to improve wellness score.')
		RETURNING id, owner_id, device_id, name, species, breed, gender, birth_date, daily_feed_target_grams, created_at
	`

	pet, err := scanPetDetailsRow(tx.QueryRow(
		ctx,
		query,
		ownerID,
		input.DeviceID,
		input.Name,
		input.Species,
		input.Breed,
		input.Gender,
		input.BirthDate,
		input.DailyFeedTargetGrams,
	))
	if err != nil {
		return domain.PetDetails{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PetDetails{}, err
	}
	return pet, nil
}

func (r *ProfileRepository) UpdatePetDetails(ctx context.Context, ownerID int64, input domain.PetDetailsInput) (domain.PetDetails, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PetDetails{}, err
	}
	defer tx.Rollback(ctx)

	if err := ensureDevice(ctx, tx, input.DeviceID); err != nil {
		return domain.PetDetails{}, err
	}

	const query = `
		UPDATE pets
		SET
			device_id = $2,
			name = $3,
			species = $4,
			breed = $5,
			gender = $6,
			birth_date = $7,
			daily_feed_target_grams = $8
		WHERE owner_id = $1
		RETURNING id, owner_id, device_id, name, species, breed, gender, birth_date, daily_feed_target_grams, created_at
	`

	pet, err := scanPetDetailsRow(tx.QueryRow(
		ctx,
		query,
		ownerID,
		input.DeviceID,
		input.Name,
		input.Species,
		input.Breed,
		input.Gender,
		input.BirthDate,
		input.DailyFeedTargetGrams,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PetDetails{}, domain.ErrNotFound
		}
		return domain.PetDetails{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PetDetails{}, err
	}
	return pet, nil
}

func (r *ProfileRepository) DeletePetDetails(ctx context.Context, ownerID int64) error {
	const query = `
		DELETE FROM pets
		WHERE owner_id = $1
	`
	tag, err := r.pool.Exec(ctx, query, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProfileRepository) GetDeviceSettings(ctx context.Context, ownerID int64) (domain.DeviceSettings, error) {
	const query = `
		SELECT
			devices.id,
			devices.name,
			devices.manual_feed_portion_grams,
			devices.food_stock_percent,
			devices.water_available,
			devices.water_status,
			devices.calibration_status,
			devices.calibration_requested_at,
			devices.last_seen_at
		FROM pets
		JOIN devices ON devices.id = pets.device_id
		WHERE pets.owner_id = $1
		ORDER BY pets.id
		LIMIT 1
	`
	return r.scanDeviceSettings(ctx, query, ownerID)
}

func (r *ProfileRepository) UpdateDeviceSettings(ctx context.Context, ownerID int64, input domain.DeviceSettingsInput) (domain.DeviceSettings, error) {
	const query = `
		UPDATE devices
		SET
			name = COALESCE(NULLIF($2, ''), devices.name),
			manual_feed_portion_grams = CASE WHEN $3::double precision > 0 THEN $3::double precision ELSE devices.manual_feed_portion_grams END
		FROM pets
		WHERE pets.device_id = devices.id
		  AND pets.owner_id = $1
		RETURNING
			devices.id,
			devices.name,
			devices.manual_feed_portion_grams,
			devices.food_stock_percent,
			devices.water_available,
			devices.water_status,
			devices.calibration_status,
			devices.calibration_requested_at,
			devices.last_seen_at
	`
	return r.scanDeviceSettings(ctx, query, ownerID, input.Name, input.ManualFeedPortionGrams)
}

func (r *ProfileRepository) CalibrateDevice(ctx context.Context, ownerID int64) (domain.DeviceSettings, error) {
	const query = `
		UPDATE devices
		SET
			calibration_status = 'tare_requested',
			calibration_requested_at = NOW()
		FROM pets
		WHERE pets.device_id = devices.id
		  AND pets.owner_id = $1
		RETURNING
			devices.id,
			devices.name,
			devices.manual_feed_portion_grams,
			devices.food_stock_percent,
			devices.water_available,
			devices.water_status,
			devices.calibration_status,
			devices.calibration_requested_at,
			devices.last_seen_at
	`
	return r.scanDeviceSettings(ctx, query, ownerID)
}

func (r *ProfileRepository) GetNotificationPreferences(ctx context.Context, ownerID int64) (domain.NotificationPreferences, error) {
	const query = `
		SELECT owner_id, low_food_alert, empty_water_alert, feeding_success_report, updated_at
		FROM notification_preferences
		WHERE owner_id = $1
	`
	return r.scanNotificationPreferences(ctx, query, ownerID)
}

func (r *ProfileRepository) UpsertNotificationPreferences(ctx context.Context, ownerID int64, input domain.NotificationPreferencesInput) (domain.NotificationPreferences, error) {
	const query = `
		INSERT INTO notification_preferences (owner_id, low_food_alert, empty_water_alert, feeding_success_report)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (owner_id) DO UPDATE SET
			low_food_alert = EXCLUDED.low_food_alert,
			empty_water_alert = EXCLUDED.empty_water_alert,
			feeding_success_report = EXCLUDED.feeding_success_report,
			updated_at = NOW()
		RETURNING owner_id, low_food_alert, empty_water_alert, feeding_success_report, updated_at
	`
	return r.scanNotificationPreferences(
		ctx,
		query,
		ownerID,
		input.LowFoodAlert,
		input.EmptyWaterAlert,
		input.FeedingSuccessReport,
	)
}

func (r *ProfileRepository) scanPetDetails(ctx context.Context, query string, args ...any) (domain.PetDetails, error) {
	pet, err := scanPetDetailsRow(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PetDetails{}, domain.ErrNotFound
		}
		return domain.PetDetails{}, err
	}
	return pet, nil
}

func (r *ProfileRepository) scanDeviceSettings(ctx context.Context, query string, args ...any) (domain.DeviceSettings, error) {
	settings, err := scanDeviceSettingsRow(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DeviceSettings{}, domain.ErrNotFound
		}
		return domain.DeviceSettings{}, err
	}
	return settings, nil
}

func (r *ProfileRepository) scanNotificationPreferences(ctx context.Context, query string, args ...any) (domain.NotificationPreferences, error) {
	var preferences domain.NotificationPreferences
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&preferences.OwnerID,
		&preferences.LowFoodAlert,
		&preferences.EmptyWaterAlert,
		&preferences.FeedingSuccessReport,
		&preferences.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotificationPreferences{}, domain.ErrNotFound
		}
		return domain.NotificationPreferences{}, err
	}
	return preferences, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPetDetailsRow(row rowScanner) (domain.PetDetails, error) {
	var pet domain.PetDetails
	var birthDate pgtype.Date
	if err := row.Scan(
		&pet.ID,
		&pet.OwnerID,
		&pet.DeviceID,
		&pet.Name,
		&pet.Species,
		&pet.Breed,
		&pet.Gender,
		&birthDate,
		&pet.DailyFeedTargetGrams,
		&pet.CreatedAt,
	); err != nil {
		return domain.PetDetails{}, err
	}
	if birthDate.Valid {
		value := birthDate.Time
		pet.BirthDate = &value
	}
	return pet, nil
}

func scanDeviceSettingsRow(row rowScanner) (domain.DeviceSettings, error) {
	var settings domain.DeviceSettings
	var calibrationRequestedAt pgtype.Timestamptz
	if err := row.Scan(
		&settings.ID,
		&settings.Name,
		&settings.ManualFeedPortionGrams,
		&settings.FoodStockPercent,
		&settings.WaterAvailable,
		&settings.WaterStatus,
		&settings.CalibrationStatus,
		&calibrationRequestedAt,
		&settings.LastSeenAt,
	); err != nil {
		return domain.DeviceSettings{}, err
	}
	if calibrationRequestedAt.Valid {
		value := calibrationRequestedAt.Time
		settings.CalibrationRequestedAt = &value
	}
	return settings, nil
}

func ensureDevice(ctx context.Context, tx pgx.Tx, deviceID string) error {
	const query = `
		INSERT INTO devices (id)
		VALUES ($1)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := tx.Exec(ctx, query, deviceID)
	return err
}
