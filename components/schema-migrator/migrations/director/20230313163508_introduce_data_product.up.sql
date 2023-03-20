BEGIN;

CREATE TABLE data_products (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id uuid NOT NULL,
    CONSTRAINT data_products_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    ord_id VARCHAR(256) NOT NULL,
    local_id VARCHAR(256),
    title VARCHAR(256),
    short_description VARCHAR(256),
    description TEXT,
    version VARCHAR(256),
    release_status release_status,
    visibility TEXT,
    CONSTRAINT data_product_visibility_check CHECK (visibility in ('public', 'internal', 'private')),
    package_id uuid NOT NULL,
    CONSTRAINT data_products_package_id_fk FOREIGN KEY (package_id) REFERENCES packages(id),
    tags JSONB,
    industry JSONB,
    line_of_business JSONB,
    product_type VARCHAR(256),
    data_product_owner VARCHAR(256)
);

ALTER TABLE bundle_references
    ADD COLUMN data_product_id uuid;

ALTER TABLE bundle_references
    ADD CONSTRAINT bundle_references_data_product_id_fkey FOREIGN KEY (data_product_id) REFERENCES data_products(id);

ALTER TABLE bundle_references
    DROP CONSTRAINT valid_refs;

CREATE TABLE ports (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    data_product_id uuid NOT NULL,
    CONSTRAINT ports_data_product_id_fk FOREIGN KEY (data_product_id) REFERENCES data_products (id) ON DELETE CASCADE,
    app_id uuid NOT NULL,
    CONSTRAINT ports_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    name VARCHAR(256),
    port_type VARCHAR(256),
    description TEXT,
    producer_cardinality VARCHAR(256),
    disabled bool DEFAULT FALSE
);

CREATE TABLE port_api_reference (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id uuid NOT NULL,
    CONSTRAINT port_api_ref_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    port_id uuid NOT NULL,
    CONSTRAINT port_api_reference_port_id_fk FOREIGN KEY (port_id) REFERENCES ports (id) ON DELETE CASCADE,
    api_id uuid NOT NULL,
    CONSTRAINT port_api_reference_api_id_fk FOREIGN KEY (api_id) REFERENCES api_definitions (id)
);

CREATE TABLE port_event_reference (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id uuid NOT NULL,
    CONSTRAINT port_event_ref_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    port_id uuid NOT NULL,
    CONSTRAINT port_event_reference_port_id_fk FOREIGN KEY (port_id) REFERENCES ports (id) ON DELETE CASCADE,
    event_id uuid NOT NULL,
    CONSTRAINT port_event_reference_event_id_fk FOREIGN KEY (event_id) REFERENCES event_api_definitions (id),
    min_version VARCHAR(256)
);

CREATE OR REPLACE VIEW data_product_tenants AS
SELECT dp.*, ta.tenant_id, ta.owner FROM data_products AS dp
                                             INNER JOIN tenant_applications ta ON ta.id = dp.app_id;

ALTER TABLE api_definitions
    DROP CONSTRAINT api_protocol_check;

ALTER TABLE api_definitions
    ADD CONSTRAINT api_protocol_check CHECK (api_protocol IN ('odata-v2', 'odata-v4', 'soap-inbound', 'soap-outbound', 'rest', 'sap-sql-api-v1'));


-- ALTER TABLE specifications
--     ALTER COLUMN api_spec_type TYPE VARCHAR(255);
--
-- DROP TYPE api_spec_type;
--
-- CREATE TYPE api_spec_type AS ENUM (
--     --- CMP types ---
--
--     'ODATA',
--     'OPEN_API',
--
--     --- ORD types ---
--
--     'openapi-v2',
--     'openapi-v3',
--     'raml-v1',
--     'edmx',
--     'csdl-json',
--     'wsdl-v1',
--     'wsdl-v2',
--     'sap-rfc-metadata-v1',
--     'sap-sql-api-definition-v1',
--     'custom'
--
--     );
--
-- ALTER TABLE specifications
--     ALTER COLUMN api_spec_type TYPE api_spec_type USING (api_spec_type::api_spec_type);

COMMIT;

ALTER TYPE api_spec_type ADD VALUE 'sap-sql-api-definition-v1' BEFORE 'custom';
