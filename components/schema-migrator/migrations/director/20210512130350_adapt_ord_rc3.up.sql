BEGIN;

ALTER TABLE vendors
    DROP COLUMN sap_partner,
    ADD COLUMN partners JSONB;

ALTER TABLE api_definitions
    ADD COLUMN extensible JSONB;

ALTER TABLE event_api_definitions
    ADD COLUMN extensible JSONB;

CREATE VIEW partners AS
SELECT vendors.ord_id AS vendor_id,
       elements.value AS value
FROM vendors,
     jsonb_array_elements_text(vendors.partners) AS elements;

CREATE VIEW api_definition_extensible AS
SELECT id                  AS api_definition_id,
       actions.supported   AS supported,
       actions.description AS description
FROM api_definitions,
     jsonb_to_record(api_definitions.extensible) AS actions(supported TEXT, description TEXT);

CREATE VIEW event_api_definition_extensible AS
SELECT id                  AS event_definition_id,
       actions.supported   AS supported,
       actions.description AS description
FROM event_api_definitions,
     jsonb_to_record(event_api_definitions.extensible) AS actions(supported TEXT, description TEXT);

COMMIT;
