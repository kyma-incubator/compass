BEGIN;

DROP VIEW IF EXISTS api_definition_extensible;
DROP VIEW IF EXISTS event_api_definition_extensible;

CREATE VIEW api_definition_extensible AS
SELECT id                  AS api_definition_id,
       actions.supported   AS supported,
       actions.description AS description
FROM api_definitions,
     jsonb_to_record(api_definitions.extensible) AS actions(supported TEXT, description TEXT)
WHERE actions.supported IS NOT NULL;

CREATE VIEW event_api_definition_extensible AS
SELECT id                  AS event_definition_id,
       actions.supported   AS supported,
       actions.description AS description
FROM event_api_definitions,
     jsonb_to_record(event_api_definitions.extensible) AS actions(supported TEXT, description TEXT)
WHERE actions.supported IS NOT NULL;

COMMIT;
