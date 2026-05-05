package bootstrap

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/domain"
	"smart-pet-monitoring/backend/internal/repository/postgres"
	"smart-pet-monitoring/backend/internal/security"
)

func EnsureDemoOwner(ctx context.Context, repo *postgres.OwnerRepository, name, email, password string) (domain.Owner, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Demo Owner"
	}
	if email == "" || strings.TrimSpace(password) == "" {
		return domain.Owner{}, nil
	}

	owner, err := repo.FindByEmail(ctx, email)
	if err == nil {
		owner.PasswordHash = ""
		return owner, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Owner{}, err
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return domain.Owner{}, err
	}

	owner = domain.Owner{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	}
	if err := repo.Create(ctx, &owner); err != nil {
		return domain.Owner{}, err
	}
	owner.PasswordHash = ""
	return owner, nil
}

func EnsureDemoProfile(ctx context.Context, db *pgxpool.Pool, owner domain.Owner) error {
	if owner.ID <= 0 {
		return nil
	}

	const deviceQuery = `
		INSERT INTO devices (id, name, food_stock_percent, water_available, water_status, last_seen_at)
		VALUES ('esp32-001', 'Kitchen Smart Feeder', 75, TRUE, 'Clean & Fresh', NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name
	`
	if _, err := db.Exec(ctx, deviceQuery); err != nil {
		return err
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
		VALUES (
			$1,
			'esp32-001',
			'Fluffy',
			'Dog',
			'Golden Retriever',
			3,
			25.4,
			150,
			92,
			'Excellent',
			'Optimal Wellness',
			'Your pet health metrics are stable this week. Keep maintaining the current diet and activity routines.',
			45,
			9.5
		)
		ON CONFLICT (owner_id) DO UPDATE SET
			device_id = EXCLUDED.device_id
		RETURNING id
	`
	var petID int64
	if err := db.QueryRow(ctx, petQuery, owner.ID).Scan(&petID); err != nil {
		return err
	}

	const taskQuery = `
		INSERT INTO care_tasks (pet_id, category, title, subtitle, due_label, due_at, priority, sort_order)
		VALUES
			($1, 'vaccination', 'Vaccination', 'Annual Rabies Booster', 'Due in 5 days', CURRENT_DATE + 5, 'high', 1),
			($1, 'checkup', 'Vet Checkup', 'General Wellness Exam', 'Oct 24', NULL, 'normal', 2)
		ON CONFLICT DO NOTHING
	`
	_, err := db.Exec(ctx, taskQuery, petID)
	return err
}
