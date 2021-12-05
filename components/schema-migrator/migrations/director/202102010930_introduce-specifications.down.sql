BEGIN;

ALTER TABLE api_definitions
    ADD COLUMN spec_data   TEXT,
    ADD COLUMN spec_format api_spec_format,
    ADD COLUMN spec_type   api_spec_type,
    ADD COLUMN api_definitions JSONB;

ALTER TABLE event_api_definitions
    ADD COLUMN spec_data   TEXT,
    ADD COLUMN spec_format event_api_spec_format,
    ADD COLUMN spec_type   event_api_spec_type,
    ADD COLUMN event_definitions JSONB;

ALTER TABLE fetch_requests
    DROP CONSTRAINT valid_refs;

DROP INDEX IF EXISTS fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx;

ALTER TABLE fetch_requests
    ADD COLUMN api_def_id       UUID,
    ADD COLUMN event_api_def_id UUID;

ALTER TABLE fetch_requests
    ADD CONSTRAINT api_def_fk FOREIGN KEY (tenant_id, api_def_id) references api_definitions (tenant_id, id) ON DELETE CASCADE;

ALTER TABLE fetch_requests
    ADD CONSTRAINT event_api_def_fk FOREIGN KEY (tenant_id, event_api_def_id) REFERENCES event_api_definitions (tenant_id, id) ON DELETE CASCADE;

INSERT INTO fetch_requests (id, tenant_id, document_id, url, auth, mode, filter, status_condition, status_timestamp, status_message, spec_id, api_def_id, event_api_def_id)
    (SELECT uuid_generate_v4(),
            f.tenant_id,
            NULL::UUID,
            f.url,
            f.auth,
            f.mode,
            f.filter,
            f.status_condition,
            f.status_timestamp,
            f.status_message,
            NULL::UUID,
            s.api_def_id,
            s.event_def_id
     FROM fetch_requests f
              JOIN specifications s ON f.tenant_id = s.tenant_id AND f.spec_id = s.id);

DELETE
FROM fetch_requests
WHERE spec_id IS NOT NULL;

ALTER TABLE fetch_requests
    DROP COLUMN spec_id;

ALTER TABLE fetch_requests
    ADD CONSTRAINT valid_refs CHECK (api_def_id IS NOT NULL OR event_api_def_id IS NOT NULL OR document_id IS NOT NULL);

CREATE UNIQUE INDEX fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx ON fetch_requests (tenant_id, coalesce(api_def_id, '00000000-0000-0000-0000-000000000000'),
                                                                                                 coalesce(event_api_def_id, '00000000-0000-0000-0000-000000000000'),
                                                                                                 coalesce(document_id, '00000000-0000-0000-0000-000000000000'));

UPDATE api_definitions
SET spec_data   = (SELECT spec_data FROM specifications WHERE api_def_id = api_definitions.id),
    spec_format = (SELECT spec_format FROM specifications WHERE api_def_id = api_definitions.id),
    spec_type   = (SELECT spec_type FROM specifications WHERE api_def_id = api_definitions.id);

UPDATE event_api_definitions
SET spec_data   = (SELECT spec_data FROM specifications WHERE event_def_id = event_api_definitions.id),
    spec_format = (SELECT spec_format FROM specifications WHERE event_def_id = event_api_definitions.id),
    spec_type   = (SELECT spec_type FROM specifications WHERE event_def_id = event_api_definitions.id);

DROP VIEW api_resource_definitions;
DROP VIEW event_resource_definitions;

DROP TABLE specifications;

ALTER TABLE vendors
    DROP CONSTRAINT vendors_tenant_id_fkey_cascade;

ALTER TABLE packages
    DROP CONSTRAINT packages_tenant_id_fkey_cascade;

ALTER TABLE products
    DROP CONSTRAINT products_tenant_id_fkey_cascade;

ALTER TABLE tombstones
    DROP CONSTRAINT tombstones_tenant_id_fkey_cascade;

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_type TYPE VARCHAR(255);

DROP TYPE event_api_spec_type;

CREATE TYPE event_api_spec_type AS ENUM (
    'ASYNC_API'
    );

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_type TYPE event_api_spec_type USING (spec_type::event_api_spec_type);

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_format TYPE VARCHAR(255);

DROP TYPE event_api_spec_format;

CREATE TYPE event_api_spec_format AS ENUM (
    'YAML',
    'XML',
    'JSON'
    );

ALTER TABLE event_api_definitions
    ALTER COLUMN spec_format TYPE event_api_spec_format USING (spec_format::event_api_spec_format);

ALTER TABLE api_definitions
    ALTER COLUMN spec_type TYPE VARCHAR(255);

DROP TYPE api_spec_type;

CREATE TYPE api_spec_type AS ENUM (
    'ODATA',
    'OPEN_API'
    );

ALTER TABLE api_definitions
    ALTER COLUMN spec_type TYPE api_spec_type USING (spec_type::api_spec_type);

ALTER TABLE api_definitions
    ALTER COLUMN spec_format TYPE VARCHAR(255);

DROP TYPE api_spec_format;

CREATE TYPE api_spec_format AS ENUM (
    'YAML',
    'XML',
    'JSON'
    );

ALTER TABLE api_definitions
    ALTER COLUMN spec_format TYPE api_spec_format USING (spec_format::api_spec_format);

CREATE VIEW api_resource_definitions AS
SELECT *
FROM (SELECT id                                  AS api_definition_id,
             api_res_defs.type                   AS type,
             api_res_defs."customType"           AS custom_type,
             format('/api/%s/specification', id) AS url,
             api_res_defs."mediaType"            AS media_type
      FROM api_definitions,
           jsonb_to_recordset(api_definitions.api_definitions) AS api_res_defs(type TEXT, "customType" TEXT,
                                                                               "mediaType" TEXT,
                                                                               url TEXT)) as api_defs
UNION ALL
(SELECT id                                  AS api_definition_id,
        CASE
            WHEN spec_type::text = 'ODATA' THEN 'edmx'
            WHEN spec_type::text = 'OPEN_API' THEN 'openapi-v3'
            ELSE spec_type::text
            END                             AS type,
        NULL                                AS custom_type,
        format('/api/%s/specification', id) AS url,
        CASE
            WHEN spec_format::text = 'YAML' THEN 'text/yaml'
            WHEN spec_format::text = 'XML' THEN 'application/xml'
            WHEN spec_format::text = 'JSON' THEN 'application/json'
            ELSE spec_format::text
            END                             AS media_type
 FROM api_definitions);

CREATE VIEW event_resource_definitions AS
SELECT *
FROM (SELECT id                                    AS event_definition_id,
             event_res_defs.type                   AS type,
             event_res_defs."customType"           AS custom_type,
             format('/event/%s/specification', id) AS url,
             event_res_defs."mediaType"            AS media_type
      FROM event_api_definitions,
           jsonb_to_recordset(event_api_definitions.event_definitions) AS event_res_defs(type TEXT, "customType" TEXT,
                                                                                         "mediaType" TEXT,
                                                                                         url TEXT)) as event_defs
UNION ALL
(SELECT id                                  AS api_definition_id,
        CASE
            WHEN spec_type::text = 'ASYNC_API' THEN 'asyncapi-v2'
            ELSE spec_type::text
            END                             AS type,
        NULL                                AS custom_type,
        format('/event/%s/specification', id) AS url,
        CASE
            WHEN spec_format::text = 'YAML' THEN 'text/yaml'
            WHEN spec_format::text = 'XML' THEN 'application/xml'
            WHEN spec_format::text = 'JSON' THEN 'application/json'
            ELSE spec_format::text
            END                             AS media_type
 FROM event_api_definitions);

DROP EXTENSION "uuid-ossp";

COMMIT;
