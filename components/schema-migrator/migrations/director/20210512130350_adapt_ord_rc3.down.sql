BEGIN;

DROP VIEW partners;

DROP VIEW api_definition_extensible;

DROP VIEW event_api_definition_extensible;

ALTER TABLE vendors
    ADD COLUMN sap_partner BOOLEAN,
    DROP COLUMN partners;

ALTER TABLE api_definitions
    DROP COLUMN extensible;

ALTER TABLE event_api_definitions
    DROP COLUMN extensible;

COMMIT;
