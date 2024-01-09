BEGIN;

ALTER table app_templates
    DROP created_at,
    DROP updated_at;

COMMIT;