ALTER TABLE pets
    ADD COLUMN IF NOT EXISTS gender TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS birth_date DATE;

ALTER TABLE devices
    ADD COLUMN IF NOT EXISTS manual_feed_portion_grams DOUBLE PRECISION NOT NULL DEFAULT 10
        CHECK (manual_feed_portion_grams >= 0),
    ADD COLUMN IF NOT EXISTS calibration_status TEXT NOT NULL DEFAULT 'idle',
    ADD COLUMN IF NOT EXISTS calibration_requested_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS pet_vital_signs (
    id BIGSERIAL PRIMARY KEY,
    pet_id BIGINT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    weight_kg DOUBLE PRECISION NOT NULL CHECK (weight_kg >= 0),
    activity_minutes INTEGER NOT NULL CHECK (activity_minutes >= 0),
    sleep_hours DOUBLE PRECISION NOT NULL CHECK (sleep_hours >= 0),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pet_vital_signs_pet_recorded_at
    ON pet_vital_signs (pet_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS notification_preferences (
    owner_id BIGINT PRIMARY KEY REFERENCES owners(id) ON DELETE CASCADE,
    low_food_alert BOOLEAN NOT NULL DEFAULT TRUE,
    empty_water_alert BOOLEAN NOT NULL DEFAULT TRUE,
    feeding_success_report BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
