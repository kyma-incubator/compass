BEGIN;

ALTER TABLE formations
    DROP COLUMN state,
    DROP COLUMN error;

COMMIT;
