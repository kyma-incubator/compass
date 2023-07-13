BEGIN;

UPDATE formation_assignments
SET value = error
WHERE error IS NOT NULL;

ALTER TABLE formation_assignments DROP COLUMN error;

COMMIT;