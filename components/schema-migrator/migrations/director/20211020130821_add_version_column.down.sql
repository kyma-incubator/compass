BEGIN;

ALTER TABLE label_definitions
    DROP COLUMN version;

ALTER TABLE labels
    DROP COLUMN version;

COMMIT;