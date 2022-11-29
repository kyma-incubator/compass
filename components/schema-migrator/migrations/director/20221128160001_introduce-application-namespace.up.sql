BEGIN;

ALTER TABLE applications ADD COLUMN application_namespace VARCHAR(256);

COMMIT;
