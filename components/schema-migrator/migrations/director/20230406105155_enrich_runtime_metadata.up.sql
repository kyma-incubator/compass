BEGIN;

ALTER TABLE runtimes
ADD COLUMN application_namespace VARCHAR(256);

COMMIT;
