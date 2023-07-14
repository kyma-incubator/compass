BEGIN;

ALTER TABLE formation_assignments ADD COLUMN error JSONB;

UPDATE formation_assignments
SET error = value
WHERE value -> 'error' IS NOT NULL;

UPDATE formation_assignments
SET value = NULL
WHERE value -> 'error' IS NOT NULL;

COMMIT;
