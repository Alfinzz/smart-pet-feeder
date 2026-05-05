CREATE TABLE IF NOT EXISTS owners (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS feed_readings (
    id BIGSERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    weight_grams DOUBLE PRECISION NOT NULL CHECK (weight_grams >= 0),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_readings_recorded_at
    ON feed_readings (recorded_at DESC);

CREATE INDEX IF NOT EXISTS idx_feed_readings_device_recorded_at
    ON feed_readings (device_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS manual_commands (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('feed', 'drink')),
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'sent', 'completed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manual_commands_device_status_created_at
    ON manual_commands (device_id, status, created_at DESC);
