BEGIN;

ALTER TABLE api_definitions
    ADD COLUMN successors JSONB;

UPDATE api_definitions
SET successors = json_build_array(to_jsonb(successor))
WHERE successor IS NOT NULL;

ALTER TABLE event_api_definitions
    ADD COLUMN successors JSONB;

UPDATE event_api_definitions
SET successors = json_build_array(to_jsonb(successor))
WHERE successor IS NOT NULL;

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
    ALTER COLUMN policy_level TYPE TEXT;

UPDATE packages
SET policy_level = 'sap:core:v1'
WHERE policy_level = 'sap';

UPDATE packages
SET policy_level = 'sap:partner:v1'
WHERE policy_level = 'sap-partner';

ALTER TABLE packages
    ALTER COLUMN policy_level TYPE policy_level
        USING policy_level::text::policy_level;

DROP TYPE policy_level_old;

ALTER TABLE api_definitions
    DROP COLUMN successor;

ALTER TABLE event_api_definitions
    DROP COLUMN successor;

COMMIT;
