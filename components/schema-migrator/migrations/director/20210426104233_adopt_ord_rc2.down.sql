BEGIN;

DROP VIEW correlation_ids;
DROP VIEW target_urls;
DROP VIEW api_bundle_reference;
DROP VIEW event_bundle_reference;

ALTER TABLE api_definitions
    DROP COLUMN implementation_standard,
    DROP COLUMN custom_implementation_standard,
    DROP COLUMN custom_implementation_standard_description;

ALTER TABLE products
    ADD COLUMN sap_ppms_object_id VARCHAR(256);

ALTER TABLE products
DROP COLUMN correlation_ids;

ALTER TABLE applications
DROP COLUMN correlation_ids;

ALTER TABLE vendors
    ADD COLUMN type vendor_type;
ALTER TABLE vendors
    ALTER COLUMN type SET NOT NULL ;

ALTER TABLE vendors
DROP COLUMN sap_partner;

-- Helper table to convert JSON array to string
CREATE TABLE convert_json_to_string_temp (
    id UUID,
    targetURL VARCHAR(256)
);

INSERT INTO convert_json_to_string_temp
    (SELECT id, jsonb_array_elements_text(target_urls) FROM api_definitions);

ALTER TABLE api_definitions
    ADD COLUMN target_url VARCHAR(256);

UPDATE api_definitions
SET target_url = temp.targetURL
FROM convert_json_to_string_temp temp
WHERE api_definitions.id = temp.id;

DROP TABLE convert_json_to_string_temp;

ALTER TABLE api_definitions
    DROP COLUMN target_urls;


-- APIs
ALTER TABLE api_definitions
DROP CONSTRAINT api_definitions_bundles_ready_fk;

ALTER TABLE event_api_definitions
DROP CONSTRAINT event_api_definitions_bundles_ready_fk;

ALTER TABLE api_definitions
    ADD COLUMN bundle_id UUID;

UPDATE api_definitions
SET bundle_id = bundle_references.bundle_id
    FROM bundle_references
WHERE api_definitions.id = bundle_references.api_def_id;

ALTER TABLE api_definitions
    ALTER COLUMN target_url SET NOT NULL;

-- Events
ALTER TABLE event_api_definitions
    ADD COLUMN bundle_id UUID;

UPDATE event_api_definitions
SET bundle_id = bundle_references.bundle_id
    FROM bundle_references
WHERE event_api_definitions.id = bundle_references.event_def_id;


ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_tenant_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id);

ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_bundles_ready_fk
        FOREIGN KEY (bundle_id, ready) REFERENCES bundles (id, ready) ON UPDATE CASCADE;

ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_tenant_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id);

ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_bundles_ready_fk
        FOREIGN KEY (bundle_id, ready) REFERENCES bundles (id, ready) ON UPDATE CASCADE;

DROP TABLE bundle_references;

COMMIT;
