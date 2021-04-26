BEGIN;

ALTER TABLE api_definitions
    ADD COLUMN implementation_standard                        VARCHAR(256),
    ADD COLUMN custom_implementation_standard                 VARCHAR(256),
    ADD COLUMN custom_implementation_standard_description     VARCHAR(255);

ALTER TABLE products
DROP COLUMN sap_ppms_object_id;
ALTER TABLE products
    ADD COLUMN correlation_ids JSONB;

ALTER TABLE applications
    ADD COLUMN correlation_ids JSONB;

ALTER TABLE vendors
DROP COLUMN type;
ALTER TABLE vendors
    ADD COLUMN sap_partner BOOLEAN;

CREATE VIEW correlation_ids AS
SELECT *
FROM (SELECT applications.id            AS application_id,
             NULL::varchar(256)         AS product_id,
             elements.value             AS value
      FROM applications,
          jsonb_array_elements_text(applications.correlation_ids) AS elements) AS app_correlation_ids
UNION ALL
(SELECT NULL::uuid         AS application_id,
        products.ord_id    AS product_id,
        elements.value     AS value
 FROM products,
     jsonb_array_elements_text(products.correlation_ids) AS elements);


ALTER TABLE api_definitions
    ADD COLUMN target_urls JSONB;

UPDATE api_definitions SET
    target_urls = json_build_array(to_jsonb(target_url))
WHERE id IS NOT NULL;

ALTER TABLE api_definitions
    ALTER COLUMN target_urls SET NOT NULL;

CREATE VIEW target_urls AS
SELECT api_definitions.id    AS api_definition_id,
       elements.value        AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.target_urls) AS elements;


-- Introduce many-to-many relation between bundles and apis/events
CREATE TABLE bundle_references (
    tenant_id UUID NOT NULL,
    api_def_id UUID,
    event_def_id UUID,
    bundle_id UUID,
    api_def_url varchar(256),
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    FOREIGN KEY (api_def_id) REFERENCES api_definitions (id) ON DELETE CASCADE,
    FOREIGN KEY (event_def_id) REFERENCES event_api_definitions (id) ON DELETE CASCADE,
    FOREIGN KEY (bundle_id) REFERENCES bundles (id) ON DELETE CASCADE,
    CONSTRAINT valid_refs CHECK ((bundle_id IS NOT NULL AND api_def_id IS NOT NULL AND api_def_url IS NOT NULL) OR (bundle_id IS NOT NULL))
);

INSERT INTO bundle_references (
    SELECT tenant_id,
           id,
           NULL::UUID,
           bundle_id,
           target_url
    FROM api_definitions
);

INSERT INTO bundle_references (
    SELECT tenant_id,
           NULL::UUID,
            id,
           bundle_id,
           NULL::varchar
    FROM event_api_definitions
);

ALTER TABLE api_definitions
DROP COLUMN target_url;

ALTER TABLE api_definitions
DROP COLUMN bundle_id;

ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_bundles_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON UPDATE CASCADE;

ALTER TABLE event_api_definitions
DROP COLUMN bundle_id;

ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_bundles_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON UPDATE CASCADE;

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

COMMIT;
