BEGIN;

ALTER TABLE app_templates DROP COLUMN application_namespace;

COMMIT;
