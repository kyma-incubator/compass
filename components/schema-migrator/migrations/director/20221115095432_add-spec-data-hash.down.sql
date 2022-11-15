BEGIN;

ALTER TABLE specifications DROP COLUMN spec_data_hash;

COMMIT;
