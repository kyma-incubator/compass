BEGIN;

ALTER TABLE applications
    DROP CONSTRAINT IF EXISTS applications_app_template_id_fkey;

COMMIT;