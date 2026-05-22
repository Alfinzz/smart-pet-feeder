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
        CHECK (action IN ('feed', 'drink', 'servo_test', 'tare'));
