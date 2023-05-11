BEGIN;

ALTER TABLE destinations
DROP COLUMN formation_assignment_id;

COMMIT;
