BEGIN;

ALTER TABLE formation_constraints
    ALTER COLUMN created_at DROP NOT NULL;

COMMIT;
