BEGIN;

ALTER TABLE formation_constraints
    ALTER COLUMN created_at SET NOT NULL;

COMMIT;
