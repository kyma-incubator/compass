BEGIN;

ALTER TABLE formation_templates
DROP COLUMN supports_reset;

COMMIT;
