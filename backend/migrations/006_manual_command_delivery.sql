ALTER TABLE manual_commands
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    ADD COLUMN IF NOT EXISTS last_error TEXT NOT NULL DEFAULT '';

UPDATE manual_commands
SET updated_at = COALESCE(updated_at, created_at, NOW());

CREATE INDEX IF NOT EXISTS idx_manual_commands_device_status_updated_at
    ON manual_commands (device_id, status, updated_at);
