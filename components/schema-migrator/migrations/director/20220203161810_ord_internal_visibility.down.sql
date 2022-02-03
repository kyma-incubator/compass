BEGIN;

ALTER TABLE bundle_references
DROP COLUMN visibility;

COMMIT;
