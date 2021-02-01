BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP VIEW api_resource_definitions;
DROP VIEW event_resource_definitions;

---------------------------- API Format -----------------------------

ALTER TABLE api_definitions
    ALTER COLUMN spec_format TYPE VARCHAR(255);

DROP TYPE api_spec_format;

CREATE TYPE api_spec_format AS ENUM (
    --- CMP formats ---

    'YAML',
    'XML',
    'JSON',

    --- ORD formats ---

    'application/json',
    'text/yaml',
    'application/xml',
    'text/plain',
    'application/octet-stream'

    );

ALTER TABLE api_definitions
    ALTER COLUMN spec_format TYPE api_spec_format USING (spec_format::api_spec_format);

---------------------------- API Type -----------------------------

ALTER TABLE api_definitions
    ALTER COLUMN spec_type TYPE VARCHAR(255);

DROP TYPE api_spec_type;

CREATE TYPE api_spec_type AS ENUM (
    --- CMP types ---

    'ODATA',
    'OPEN_API',

    --- ORD types ---

    'openapi-v2',
    'openapi-v3',
    'raml-v1',
    'edmx',
    'csdl-json',
    'wsdl-v1',
    'wsdl-v2',
    'sap-rfc-metadata-v1',
    'custom'

    );

ALTER TABLE api_definitions
    ALTER COLUMN spec_type TYPE api_spec_type USING (spec_type::api_spec_type);

---------------------------- Event Format -----------------------------

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_format TYPE VARCHAR(255);

DROP TYPE event_api_spec_format;

CREATE TYPE event_api_spec_format AS ENUM (
    --- CMP formats ---

    'YAML',
    'XML',
    'JSON',

    --- ORD formats ---

    'application/json',
    'text/yaml',
    'application/xml',
    'text/plain',
    'application/octet-stream'

    );

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_format TYPE event_api_spec_format USING (spec_format::event_api_spec_format);

---------------------------- Event Type -----------------------------

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_type TYPE VARCHAR(255);

DROP TYPE event_api_spec_type;

CREATE TYPE event_api_spec_type AS ENUM (
    --- CMP types ---

    'ASYNC_API',

    --- ORD types ---

    'asyncapi-v2',
    'custom'

    );

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_type TYPE event_api_spec_type USING (spec_type::event_api_spec_type);

-------------- Orphan Data Deletion for ORD resources ---------------

ALTER TABLE vendors
    ADD CONSTRAINT vendors_tenant_id_fkey_cascade
        FOREIGN KEY (tenant_id)
            REFERENCES business_tenant_mappings (id)
            ON DELETE CASCADE;

ALTER TABLE packages
    ADD CONSTRAINT packages_tenant_id_fkey_cascade
        FOREIGN KEY (tenant_id)
            REFERENCES business_tenant_mappings (id)
            ON DELETE CASCADE;

ALTER TABLE products
    ADD CONSTRAINT products_tenant_id_fkey_cascade
        FOREIGN KEY (tenant_id)
            REFERENCES business_tenant_mappings (id)
            ON DELETE CASCADE;

ALTER TABLE tombstones
    ADD CONSTRAINT tombstones_tenant_id_fkey_cascade
        FOREIGN KEY (tenant_id)
            REFERENCES business_tenant_mappings (id)
            ON DELETE CASCADE;

---------------------------------------------------------------------

CREATE TABLE specifications
(
    id                UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id         UUID NOT NULL,
    api_def_id        UUID,
    event_def_id      UUID,
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    FOREIGN KEY (api_def_id) REFERENCES api_definitions (id) ON DELETE CASCADE,
    FOREIGN KEY (event_def_id) REFERENCES event_api_definitions (id) ON DELETE CASCADE,
    spec_data         TEXT,
    api_spec_format   api_spec_format,
    api_spec_type     api_spec_type,
    event_spec_format event_api_spec_format,
    event_spec_type   event_api_spec_type,
    custom_type       VARCHAR(256)
        CONSTRAINT valid_refs CHECK ((api_def_id IS NOT NULL AND api_spec_format IS NOT NULL AND
                                      api_spec_type IS NOT NULL) OR
                                     (event_def_id IS NOT NULL AND event_spec_format IS NOT NULL AND
                                      event_spec_type IS NOT NULL))
);

CREATE INDEX ON specifications (tenant_id);
CREATE UNIQUE INDEX ON specifications (tenant_id, coalesce(api_def_id, '00000000-0000-0000-0000-000000000000'),
                                       coalesce(event_def_id, '00000000-0000-0000-0000-000000000000'));
CREATE UNIQUE INDEX ON specifications (tenant_id, id);

INSERT INTO specifications
    (SELECT uuid_generate_v4(),
            tenant_id,
            id,
            NULL::UUID,
            spec_data,
            spec_format,
            spec_type,
            NULL::event_api_spec_format,
            NULL::event_api_spec_type,
            NULL
     FROM api_definitions);

INSERT INTO specifications
    (SELECT uuid_generate_v4(),
            tenant_id,
            NULL::UUID,
            id,
            spec_data,
            NULL::api_spec_format,
            NULL::api_spec_type,
            spec_format,
            spec_type,
            NULL
     FROM event_api_definitions);

ALTER TABLE fetch_requests
    ADD COLUMN spec_id UUID;

ALTER TABLE fetch_requests
    ADD CONSTRAINT spec_id_fk FOREIGN KEY (spec_id) REFERENCES specifications (id) ON DELETE CASCADE;

ALTER TABLE fetch_requests
    DROP CONSTRAINT valid_refs;

DROP INDEX IF EXISTS fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx;

INSERT INTO fetch_requests
    (SELECT uuid_generate_v4(),
            f.tenant_id,
            NULL::UUID,
            NULL::UUID,
            NULL::UUID,
            f.url,
            f.auth,
            f.mode,
            f.filter,
            f.status_condition,
            f.status_timestamp,
            f.status_message,
            s.id
     FROM fetch_requests f
              JOIN specifications s ON (f.tenant_id = s.tenant_id AND f.api_def_id = s.api_def_id) OR
                                       (f.tenant_id = s.tenant_id AND f.event_api_def_id = s.event_def_id));

DELETE
FROM fetch_requests
WHERE event_api_def_id IS NOT NULL
   OR api_def_id IS NOT NULL;

ALTER TABLE fetch_requests
    DROP COLUMN api_def_id,
    DROP COLUMN event_api_def_id;

ALTER TABLE fetch_requests
    ADD CONSTRAINT valid_refs
        CHECK (document_id IS NOT NULL OR spec_id IS NOT NULL);

CREATE UNIQUE INDEX fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx ON fetch_requests (tenant_id, coalesce(document_id, '00000000-0000-0000-0000-000000000000'),
                                       coalesce(spec_id, '00000000-0000-0000-0000-000000000000'));

ALTER TABLE api_definitions
    DROP COLUMN spec_data,
    DROP COLUMN spec_format,
    DROP COLUMN spec_type;

ALTER TABLE event_api_definitions
    DROP COLUMN spec_data,
    DROP COLUMN spec_format,
    DROP COLUMN spec_type;

-- TODO: Add spec views for ORD Service

COMMIT;
