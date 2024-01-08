BEGIN;

-- Drop views --

DROP VIEW IF EXISTS api_resource_definitions;
DROP VIEW IF EXISTS api_specifications_tenants;
DROP VIEW IF EXISTS event_specifications_tenants;
DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_products;
DROP VIEW IF EXISTS tenants_tombstones;
DROP VIEW IF EXISTS ord_hierarchy_event_definitions;
DROP VIEW IF EXISTS ord_hierarchy_api_definitions;


-- Alter tables --

ALTER TABLE products
    DROP COLUMN description;

ALTER TABLE tombstones
    DROP COLUMN description;

ALTER TABLE api_definitions
    ADD COLUMN hierarchy JSONB;

ALTER TABLE event_api_definitions
    ADD COLUMN hierarchy JSONB;

ALTER TABLE api_definitions
    DROP CONSTRAINT api_protocol_check;

ALTER TABLE api_definitions
    ADD CONSTRAINT api_protocol_check CHECK (api_protocol IN ('odata-v2', 'odata-v4', 'soap-inbound', 'soap-outbound', 'rest', 'websocket', 'sap-rfc', 'sap-sql-api-v1'));

ALTER TABLE specifications
    RENAME COLUMN api_spec_type TO api_spec_type_old;

ALTER TABLE specifications
    ADD COLUMN api_spec_type_var VARCHAR(255);

UPDATE specifications
    SET api_spec_type_var = api_spec_type_old;

ALTER TABLE specifications
    DROP COLUMN api_spec_type_old;

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
    ADD COLUMN api_spec_type api_spec_type;

UPDATE specifications
SET api_spec_type = CAST(api_spec_type_var AS api_spec_type);

ALTER TABLE specifications
    DROP COLUMN api_spec_type_var;


-- Create new views --

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


-- Recreate views --

CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, formation_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags, hierarchy,
             supported_use_cases, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible_supported, extensible_description, successors, resource_hash,
             documentation_labels, correlation_ids, direction)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                apis.id,
                apis.app_id,
                apis.name,
                apis.description,
                apis.group_name,
                apis.default_auth,
                apis.version_value,
                apis.version_deprecated,
                apis.version_deprecated_since,
                apis.version_for_removal,
                apis.ord_id,
                apis.local_tenant_id,
                apis.short_description,
                apis.system_instance_aware,
                apis.policy_level,
                apis.custom_policy_level,
                apis.api_protocol,
                apis.tags,
                apis.hierarchy,
                apis.supported_use_cases,
                apis.countries,
                apis.links,
                apis.api_resource_links,
                apis.release_status,
                apis.sunset_date,
                apis.changelog_entries,
                apis.labels,
                apis.package_id,
                apis.visibility,
                apis.disabled,
                apis.part_of_products,
                apis.line_of_business,
                apis.industry,
                apis.ready,
                apis.created_at,
                apis.updated_at,
                apis.deleted_at,
                apis.error,
                apis.implementation_standard,
                apis.custom_implementation_standard,
                apis.custom_implementation_standard_description,
                apis.target_urls,
                actions.supported,
                actions.description,
                apis.successors,
                apis.resource_hash,
                apis.documentation_labels,
                apis.correlation_ids,
                apis.direction
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id,
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
               FROM apps_subaccounts) t_apps ON apis.app_id = t_apps.id,
     -- breaking down the extensible field; the new fields will be extensible_supported and extensible_description
     jsonb_to_record(apis.extensible) actions(supported text, description text);


CREATE OR REPLACE VIEW tenants_events
            (tenant_id, formation_id, id, app_id, name, description, group_name, version_value, version_deprecated,
             version_deprecated_since, version_for_removal, ord_id, local_tenant_id, short_description,
             system_instance_aware, policy_level, custom_policy_level, changelog_entries, links, tags, hierarchy,
             countries, release_status, sunset_date, labels, package_id, visibility, disabled, part_of_products,
             line_of_business, industry, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, extensible_supported, extensible_description, successors,
             resource_hash, correlation_ids)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                events.id,
                events.app_id,
                events.name,
                events.description,
                events.group_name,
                events.version_value,
                events.version_deprecated,
                events.version_deprecated_since,
                events.version_for_removal,
                events.ord_id,
                events.local_tenant_id,
                events.short_description,
                events.system_instance_aware,
                events.policy_level,
                events.custom_policy_level,
                events.changelog_entries,
                events.links,
                events.tags,
                events.hierarchy,
                events.countries,
                events.release_status,
                events.sunset_date,
                events.labels,
                events.package_id,
                events.visibility,
                events.disabled,
                events.part_of_products,
                events.line_of_business,
                events.industry,
                events.ready,
                events.created_at,
                events.updated_at,
                events.deleted_at,
                events.error,
                events.implementation_standard,
                events.custom_implementation_standard,
                events.custom_implementation_standard_description,
                actions.supported,
                actions.description,
                events.successors,
                events.resource_hash,
                events.correlation_ids
FROM event_api_definitions events
         JOIN (SELECT a1.id,
                      a1.tenant_id,
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
               FROM apps_subaccounts) t_apps ON events.app_id = t_apps.id,
     -- breaking down the extensible field; the new fields will be extensible_supported and extensible_description
     jsonb_to_record(events.extensible) actions(supported text, description text);


CREATE OR REPLACE VIEW tenants_products
            (tenant_id, formation_id, ord_id, app_id, title, short_description, vendor, parent, labels,
             tags, correlation_ids, id, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                p.ord_id,
                p.app_id,
                p.title,
                p.short_description,
                p.vendor,
                p.parent,
                p.labels,
                p.tags,
                p.correlation_ids,
                p.id,
                p.documentation_labels
FROM products p
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
              ON p.app_id = t_apps.id OR p.app_id IS NULL;


CREATE OR REPLACE VIEW tenants_tombstones(tenant_id, formation_id, ord_id, app_id, removal_date, id)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                t.ord_id,
                t.app_id,
                t.removal_date,
                t.id
FROM tombstones t
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
              ON t.app_id = t_apps.id;


CREATE OR REPLACE VIEW api_resource_definitions AS
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


CREATE OR REPLACE VIEW api_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                             INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ad.app_id);

CREATE OR REPLACE VIEW event_specifications_tenants AS
SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                            INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                            INNER JOIN tenant_applications ta on ta.id = ead.app_id;

CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format,
             event_spec_type, custom_type, created_at)
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

COMMIT;
