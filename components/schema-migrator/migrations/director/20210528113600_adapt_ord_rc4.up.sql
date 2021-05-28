BEGIN;

ALTER TABLE api_definitions
    DROP COLUMN successor,
    ADD COLUMN successors JSONB;

ALTER TABLE event_api_definitions
    DROP COLUMN successor,
    ADD COLUMN successors JSONB;

CREATE VIEW api_definition_successors AS
SELECT id             AS api_definition_id,
       elements.value AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.successors) AS elements;

CREATE VIEW event_api_definition_successors AS
SELECT id             AS event_definition_id,
       elements.value AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.successors) AS elements;

ALTER TYPE policy_level RENAME TO policy_level_old;

CREATE TYPE policy_level AS ENUM ('custom', 'sap:core:v1','sap:partner:v1');

ALTER TABLE packages
    ALTER COLUMN policy_level TYPE policy_level
    USING policy_level::text::policy_level;

DROP TYPE policy_level_old;

COMMIT;
