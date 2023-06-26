BEGIN;

ALTER TABLE destinations
DROP COLUMN formation_assignment_id;

ALTER TABLE destinations
ALTER COLUMN revision SET NOT NULL,
ALTER COLUMN bundle_id SET NOT NULL;

COMMIT;
