BEGIN;

ALTER TABLE formation_assignments
    ALTER COLUMN source_type TYPE VARCHAR(256),
    ALTER COLUMN target_type TYPE VARCHAR(256);

ALTER TABLE formation_assignments
    DROP CONSTRAINT IF EXISTS source_type_check,
    DROP CONSTRAINT IF EXISTS target_type_check;

COMMIT;
