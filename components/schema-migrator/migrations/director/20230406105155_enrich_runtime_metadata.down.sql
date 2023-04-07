BEGIN;

ALTER TABLE runtimes
DROP COLUMN application_namespace;

COMMIT;
