BEGIN;

ALTER TABLE formation_constraints
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE formation_constraints
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

COMMIT;
