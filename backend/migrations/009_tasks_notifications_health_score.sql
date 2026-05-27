ALTER TABLE owners
    ADD COLUMN IF NOT EXISTS alert_low_food BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS alert_empty_water BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS alert_feed_success BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notification_preferences_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE owners
SET
    alert_low_food = notification_preferences.low_food_alert,
    alert_empty_water = notification_preferences.empty_water_alert,
    alert_feed_success = notification_preferences.feeding_success_report,
    notification_preferences_updated_at = notification_preferences.updated_at
FROM notification_preferences
WHERE notification_preferences.owner_id = owners.id;

ALTER TABLE care_tasks
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'completed'));

CREATE TABLE IF NOT EXISTS pet_tasks (
    id BIGSERIAL PRIMARY KEY,
    pet_id BIGINT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    category TEXT NOT NULL DEFAULT 'medical',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    due_date DATE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'completed')),
    priority TEXT NOT NULL DEFAULT 'normal',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (pet_id, title, description)
);

INSERT INTO pet_tasks (
    id,
    pet_id,
    category,
    title,
    description,
    due_date,
    status,
    priority,
    sort_order,
    created_at
)
SELECT
    id,
    pet_id,
    category,
    title,
    subtitle,
    due_at,
    status,
    priority,
    sort_order,
    created_at
FROM care_tasks
ON CONFLICT DO NOTHING;

SELECT setval(
    pg_get_serial_sequence('pet_tasks', 'id'),
    COALESCE((SELECT MAX(id) FROM pet_tasks), 1),
    (SELECT COUNT(*) > 0 FROM pet_tasks)
);

CREATE INDEX IF NOT EXISTS idx_pet_tasks_pet_status_due
    ON pet_tasks (pet_id, status, due_date);

CREATE INDEX IF NOT EXISTS idx_pet_tasks_pet_sort
    ON pet_tasks (pet_id, sort_order, due_date);
