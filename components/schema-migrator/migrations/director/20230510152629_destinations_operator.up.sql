BEGIN;

ALTER TABLE destinations
ADD COLUMN formation_assignment_id uuid REFERENCES formation_assignments(id);

COMMIT;
