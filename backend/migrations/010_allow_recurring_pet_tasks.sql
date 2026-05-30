ALTER TABLE pet_tasks
    DROP CONSTRAINT IF EXISTS pet_tasks_pet_id_title_description_key;

CREATE INDEX IF NOT EXISTS idx_pet_tasks_pet_title_description
    ON pet_tasks (pet_id, title, description);
