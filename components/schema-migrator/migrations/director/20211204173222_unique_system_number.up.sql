BEGIN;

ALTER TABLE applications ADD CONSTRAINT unique_system_number UNIQUE (system_number);

COMMIT;
