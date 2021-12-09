BEGIN;

ALTER TABLE bundle_references
ADD is_default_bundle bool;

DROP VIEW api_bundle_reference;
DROP VIEW event_bundle_reference;

CREATE VIEW api_bundle_reference AS
SELECT api_def_id        AS api_definition_id,
       api_def_url       AS default_entry_point,
       bundle_id         AS bundle_id,
       is_default_bundle AS default_consumption_bundle
FROM bundle_references
WHERE api_def_id IS NOT NULL;

CREATE VIEW event_bundle_reference AS
SELECT event_def_id           AS event_definition_id,
       NULL::VARCHAR(256)     AS default_entry_point,
       bundle_id              AS bundle_id,
       is_default_bundle      AS default_consumption_bundle
FROM bundle_references
WHERE event_def_id IS NOT NULL;

COMMIT;
