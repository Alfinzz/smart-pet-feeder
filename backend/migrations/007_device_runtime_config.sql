ALTER TABLE devices
    ADD COLUMN IF NOT EXISTS servo_open_degrees INTEGER NOT NULL DEFAULT 25
        CHECK (servo_open_degrees >= 0 AND servo_open_degrees <= 180),
    ADD COLUMN IF NOT EXISTS servo_closed_degrees INTEGER NOT NULL DEFAULT 55
        CHECK (servo_closed_degrees >= 0 AND servo_closed_degrees <= 180),
    ADD COLUMN IF NOT EXISTS automation_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS config_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE devices
SET
    servo_open_degrees = COALESCE(servo_open_degrees, 25),
    servo_closed_degrees = COALESCE(servo_closed_degrees, 55),
    automation_enabled = COALESCE(automation_enabled, FALSE),
    config_updated_at = COALESCE(config_updated_at, NOW());

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'manual_commands_action_check'
    ) THEN
        ALTER TABLE manual_commands
            DROP CONSTRAINT manual_commands_action_check;
    END IF;
END $$;

ALTER TABLE manual_commands
    ADD CONSTRAINT manual_commands_action_check
        CHECK (action IN ('feed', 'drink', 'servo_test'));
