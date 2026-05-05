CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT 'Smart Feeder',
    food_stock_percent DOUBLE PRECISION NOT NULL DEFAULT 75
        CHECK (food_stock_percent >= 0 AND food_stock_percent <= 100),
    water_available BOOLEAN NOT NULL DEFAULT TRUE,
    water_status TEXT NOT NULL DEFAULT 'Clean & Fresh',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pets (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    species TEXT NOT NULL DEFAULT 'Dog',
    breed TEXT NOT NULL DEFAULT 'Golden Retriever',
    age_years INTEGER NOT NULL DEFAULT 3 CHECK (age_years >= 0),
    weight_kg DOUBLE PRECISION NOT NULL DEFAULT 25.4 CHECK (weight_kg >= 0),
    daily_feed_target_grams DOUBLE PRECISION NOT NULL DEFAULT 150
        CHECK (daily_feed_target_grams >= 0),
    health_score INTEGER NOT NULL DEFAULT 92 CHECK (health_score >= 0 AND health_score <= 100),
    health_status TEXT NOT NULL DEFAULT 'Excellent',
    health_headline TEXT NOT NULL DEFAULT 'Optimal Wellness',
    health_description TEXT NOT NULL DEFAULT 'Your pet health metrics are stable this week. Keep maintaining the current diet and activity routines.',
    activity_minutes INTEGER NOT NULL DEFAULT 45 CHECK (activity_minutes >= 0),
    sleep_hours DOUBLE PRECISION NOT NULL DEFAULT 9.5 CHECK (sleep_hours >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id)
);

CREATE INDEX IF NOT EXISTS idx_pets_device_id
    ON pets (device_id);

CREATE TABLE IF NOT EXISTS care_tasks (
    id BIGSERIAL PRIMARY KEY,
    pet_id BIGINT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    subtitle TEXT NOT NULL,
    due_label TEXT NOT NULL,
    due_at DATE,
    priority TEXT NOT NULL DEFAULT 'normal',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (pet_id, title, subtitle)
);

CREATE INDEX IF NOT EXISTS idx_care_tasks_pet_sort
    ON care_tasks (pet_id, sort_order, due_at);
