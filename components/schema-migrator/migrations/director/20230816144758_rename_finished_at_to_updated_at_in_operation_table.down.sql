BEGIN;

ALTER TABLE operation
    RENAME COLUMN updated_at TO finished_at;

COMMIT;
