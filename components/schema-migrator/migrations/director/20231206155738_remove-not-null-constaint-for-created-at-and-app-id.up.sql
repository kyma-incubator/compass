BEGIN;

ALTER TABLE capabilities
    ALTER COLUMN created_at DROP NOT NULL;

ALTER TABLE integration_dependencies
    ALTER COLUMN app_id DROP NOT NULL,
    ALTER COLUMN created_at DROP NOT NULL;

ALTER TABLE aspects
    ALTER COLUMN app_id DROP NOT NULL;

COMMIT;
