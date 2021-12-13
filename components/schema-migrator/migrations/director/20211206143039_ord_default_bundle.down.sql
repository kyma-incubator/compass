BEGIN;

DROP VIEW api_bundle_reference;
DROP VIEW event_bundle_reference;

CREATE VIEW api_bundle_reference AS
SELECT api_def_id      AS api_definition_id,
       api_def_url     AS default_entry_point,
       bundle_id       AS bundle_id
FROM bundle_references
WHERE api_def_id IS NOT NULL;

CREATE VIEW event_bundle_reference AS
SELECT event_def_id           AS event_definition_id,
       NULL::VARCHAR(256)     AS default_entry_point,
       bundle_id              AS bundle_id
FROM bundle_references
WHERE event_def_id IS NOT NULL;

ALTER TABLE bundle_references
DROP COLUMN is_default_bundle;

COMMIT;
