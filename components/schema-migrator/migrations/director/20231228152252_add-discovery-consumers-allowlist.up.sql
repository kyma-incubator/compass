BEGIN;

ALTER TABLE formation_templates
    ADD COLUMN discovery_consumers JSONB;

CREATE INDEX discovery_consumers_gin ON formation_templates USING gin (discovery_consumers);
CREATE INDEX scenario_label_gin ON labels USING gin (value) WHERE key = 'scenarios';

COMMIT;
