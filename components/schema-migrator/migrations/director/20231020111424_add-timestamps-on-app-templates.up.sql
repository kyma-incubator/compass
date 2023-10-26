BEGIN;

ALTER table app_templates
    ADD COLUMN created_at timestamp,
    ADD COLUMN updated_at timestamp;

UPDATE app_templates SET created_at = now()
    WHERE app_templates.created_at is null;

UPDATE app_templates SET updated_at = now()
    WHERE app_templates.updated_at is null;

COMMIT;