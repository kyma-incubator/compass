BEGIN;

ALTER TABLE formation_constraints
    ALTER COLUMN created_at DROP NOT NULL;

ALTER TABLE formation_constraints
    ALTER COLUMN created_at SET DEFAULT NULL;

COMMIT;
