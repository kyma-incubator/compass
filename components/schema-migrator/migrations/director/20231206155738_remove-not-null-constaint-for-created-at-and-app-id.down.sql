BEGIN;

ALTER TABLE aspects
    ALTER COLUMN app_id SET NOT NULL;

ALTER TABLE integration_dependencies
    ALTER COLUMN app_id SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE capabilities
    ALTER COLUMN created_at SET NOT NULL;


COMMIT;
