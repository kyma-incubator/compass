BEGIN;

ALTER TABLE specifications ADD COLUMN spec_data_hash VARCHAR(256);

COMMIT;
