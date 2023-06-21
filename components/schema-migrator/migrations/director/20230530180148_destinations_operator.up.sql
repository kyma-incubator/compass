BEGIN;

ALTER TABLE destinations
ADD COLUMN formation_assignment_id uuid REFERENCES formation_assignments(id);

ALTER TABLE destinations
ALTER COLUMN revision DROP NOT NULL,
ALTER COLUMN bundle_id DROP NOT NULL;

COMMIT;
