BEGIN;

ALTER TABLE formation_templates
    ADD COLUMN discovery_consumers JSONB;

CREATE INDEX discovery_consumers_gin ON formation_templates USING gin (discovery_consumers);
CREATE INDEX scenario_label_gin ON labels USING gin (value) WHERE key = 'scenarios';

-- Initial set of discovery consumers to ensure existing scenarios will continue working
UPDATE formation_templates SET discovery_consumers = runtime_types;

COMMIT;
