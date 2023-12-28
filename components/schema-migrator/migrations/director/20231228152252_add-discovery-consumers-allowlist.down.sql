BEGIN;

ALTER TABLE formation_templates
    DROP COLUMN discovery_consumers;

DROP INDEX discovery_consumers_gin
DROP INDEX scenario_label_gin

COMMIT;
