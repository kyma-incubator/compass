BEGIN;

DROP VIEW IF EXISTS tenants_apps;
DROP VIEW IF EXISTS listening_applications;
DROP VIEW IF EXISTS api_resource_definitions;
DROP VIEW IF EXISTS api_specifications_tenants;
DROP VIEW IF EXISTS event_specifications_tenants;
DROP VIEW IF EXISTS tenants_specifications;

ALTER TABLE products
    ADD COLUMN tags JSONB;

ALTER TABLE vendors
    ADD COLUMN tags JSONB;

ALTER TABLE applications
    ADD COLUMN tags JSONB;

ALTER TABLE bundles
    ADD COLUMN tags JSONB;

ALTER TABLE api_definitions
    ADD COLUMN hierarchy JSONB;

ALTER TABLE api_definitions
    ADD COLUMN supported_use_cases JSONB;

ALTER TABLE api_definitions 
    DROP CONSTRAINT api_protocol_check;

ALTER TABLE api_definitions
    ADD CONSTRAINT api_protocol_check CHECK (api_protocol IN ('odata-v2', 'odata-v4', 'soap-inbound', 'soap-outbound', 'rest', 'websocket', 'sap-rfc', 'sap-sql-api-v1'));

ALTER TABLE event_api_definitions
    ADD COLUMN hierarchy JSONB;

ALTER TABLE event_api_definitions
    ADD COLUMN supported_use_cases JSONB;

CREATE VIEW ord_tags_products AS
SELECT id                  AS product_id,
       elements.value      AS value
FROM products,
     jsonb_array_elements_text(products.tags) AS elements;

CREATE VIEW ord_tags_vendors AS
SELECT id                  AS vendor_id,
       elements.value      AS value
FROM vendors,
     jsonb_array_elements_text(vendors.tags) AS elements;

CREATE VIEW ord_tags_applications AS
SELECT id                  AS application_id,
       elements.value      AS value
FROM applications,
     jsonb_array_elements_text(applications.tags) AS elements;

CREATE VIEW ord_tags_bundles AS
SELECT id                  AS bundle_id,
       elements.value      AS value
FROM bundles,
     jsonb_array_elements_text(bundles.tags) AS elements;

CREATE VIEW ord_hierarchy_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.hierarchy) AS elements;

CREATE VIEW ord_hierarchy_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.hierarchy) AS elements;

CREATE VIEW ord_supported_use_cases_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.supported_use_cases) AS elements;

CREATE VIEW ord_supported_use_cases_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.supported_use_cases) AS elements;

ALTER TABLE specifications
    ALTER COLUMN api_spec_type TYPE VARCHAR(255);

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
    'sap-sql-api-definition-v1',
    'custom'

    );

ALTER TABLE specifications
    ALTER COLUMN api_spec_type TYPE api_spec_type USING (api_spec_type::api_spec_type);

CREATE VIEW api_resource_definitions AS
SELECT api_def_id                                         AS api_definition_id,
       CASE
           WHEN api_spec_type::text = 'ODATA' THEN 'edmx'
           WHEN api_spec_type::text = 'OPEN_API' THEN 'openapi-v3'
           ELSE api_spec_type::text
           END                                            AS type,
       custom_type                                        AS custom_type,
       format('/api/%s/specification/%s', api_def_id, id) AS url,
       CASE
           WHEN api_spec_format::text = 'YAML' THEN 'text/yaml'
           WHEN api_spec_format::text = 'XML' THEN 'application/xml'
           WHEN api_spec_format::text = 'JSON' THEN 'application/json'
           ELSE api_spec_format::text
           END                                            AS media_type
FROM specifications
WHERE api_def_id IS NOT NULL;

CREATE VIEW api_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                             INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ad.app_id);

CREATE VIEW event_specifications_tenants AS
SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                            INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                            INNER JOIN tenant_applications ta on ta.id = ead.app_id;

CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type,
             event_spec_format, event_spec_type, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                spec.id,
                spec.api_def_id,
                spec.event_def_id,
                spec.spec_data,
                spec.api_spec_format,
                spec.api_spec_type,
                spec.event_spec_format,
                spec.event_spec_type,
                spec.custom_type,
                spec.created_at
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e) t_api_event_def
              ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;


CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, formation_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, application_namespace, product_type)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                apps.id,
                apps.name,
                apps.description,
                apps.status_condition,
                apps.status_timestamp,
                apps.healthcheck_url,
                apps.integration_system_id,
                apps.provider_name,
                apps.base_url,
                apps.labels,
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                coalesce(apps.application_namespace, tmpl.application_namespace),
                tmpl.name AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      apps_subaccounts.formation_id
               FROM apps_subaccounts
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps
              ON apps.id = t_apps.id;

CREATE
OR REPLACE VIEW listening_applications
            (id, app_template_id, system_number, local_tenant_id, name, description, status_condition, status_timestamp,
             system_status, healthcheck_url, integration_system_id, provider_name, base_url, application_namespace,
             labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, documentation_labels,
             webhook_type)
AS
SELECT a.id,
       a.app_template_id,
       a.system_number,
       a.local_tenant_id,
       a.name,
       a.description,
       a.status_condition,
       a.status_timestamp,
       a.system_status,
       a.healthcheck_url,
       a.integration_system_id,
       a.provider_name,
       a.base_url,
       a.application_namespace,
       a.labels,
       a.ready,
       a.created_at,
       a.updated_at,
       a.deleted_at,
       a.error,
       a.correlation_ids,
       a.tags,
       a.documentation_labels,
       w.type
FROM applications a
         JOIN webhooks w on w.app_id = a.id or w.app_template_id = a.app_template_id;

COMMIT;