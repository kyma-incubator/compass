BEGIN;

ALTER TABLE operation
    RENAME COLUMN finished_at TO updated_at;

COMMIT;
