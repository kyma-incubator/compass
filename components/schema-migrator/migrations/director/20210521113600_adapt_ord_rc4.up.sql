BEGIN;

ALTER TABLE api_definitions
    DROP COLUMN successor,
    ADD COLUMN successors JSONB;

ALTER TABLE event_api_definitions
    DROP COLUMN successor,
    ADD COLUMN successors JSONB;

CREATE VIEW api_definition_successors AS
SELECT id                  AS api_definition_id,
       elements.value AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.successors) AS elements;

CREATE VIEW event_api_definition_successors AS
SELECT id                  AS event_definition_id,
       elements.value AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.successors) AS elements;

COMMIT;
