BEGIN;

DROP VIEW api_definition_successors;
DROP VIEW event_api_definition_successors;

ALTER TABLE api_definitions
    ADD COLUMN successor VARCHAR(256),
    DROP COLUMN successors;

ALTER TABLE event_api_definitions
    ADD COLUMN successor VARCHAR(256),
    DROP COLUMN successors;

COMMIT;
