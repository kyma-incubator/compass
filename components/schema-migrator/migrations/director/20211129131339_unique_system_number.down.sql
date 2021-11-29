BEGIN;

ALTER TABLE applications DROP CONSTRAINT unique_system_number;

COMMIT;
