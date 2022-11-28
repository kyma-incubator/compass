BEGIN;

ALTER TABLE applications DROP COLUMN application_namespace;

COMMIT;
