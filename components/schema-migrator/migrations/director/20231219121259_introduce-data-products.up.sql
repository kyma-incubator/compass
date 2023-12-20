BEGIN;

DROP VIEW IF EXISTS tenants_specifications; -- this one won't be changed but it uses tenants_apis/events and it has to be dropped so the other two can be as well
DROP VIEW IF EXISTS tenants_entity_type_mappings; -- this one won't be changed but it uses tenants_apis/events and it has to be dropped so the other two can be as well
DROP VIEW IF EXISTS entity_type_mappings_tenants;

-- Drop and recreate tenants_apis view - add `responsible` and `usage` columns
DROP VIEW IF EXISTS tenants_apis;
-- Drop and recreate tenants_events view - add `responsible` column
DROP VIEW IF EXISTS tenants_events;

-- Drop and recreate tenants_packages view - add `runtime_restriction` column
DROP VIEW IF EXISTS tenants_packages;

-- Add `runtimeRestriction` to Package
ALTER TABLE packages
    ADD COLUMN runtime_restriction VARCHAR(256);

-- Add `responsible` and `usage` to API
ALTER TABLE api_definitions
    ADD COLUMN responsible VARCHAR(256),
    ADD COLUMN usage VARCHAR(256);

-- Add `responsible` to Event
ALTER TABLE event_api_definitions
    ADD COLUMN responsible VARCHAR(256);

-- Create data_products table
CREATE TABLE data_products
(
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id UUID NOT NULL,
        CONSTRAINT data_products_application_id_fkey FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    app_template_version_id UUID,
        CONSTRAINT data_products_app_template_version_id_fk FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
        CONSTRAINT data_products_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id),
    ord_id VARCHAR(256) NOT NULL,
        CONSTRAINT data_products_ord_id_unique UNIQUE (app_id, ord_id),
    local_tenant_id VARCHAR(256),
    correlation_ids JSONB,
    title VARCHAR(256) NOT NULL,
    short_description VARCHAR(256),
    description TEXT,
    package_id UUID,
        CONSTRAINT data_products_package_id_fk FOREIGN KEY (package_id) REFERENCES packages (id) ON DELETE CASCADE,
    last_update VARCHAR(256),
    visibility TEXT NOT NULL,
    release_status release_status NOT NULL,
    disabled BOOLEAN,
    deprecation_date VARCHAR(256),
    sunset_date VARCHAR(256),
    successors JSONB,
    changelog_entries JSONB,
    type VARCHAR(256),
    category VARCHAR(256),
    entity_types JSONB,
    input_ports JSONB,
    output_ports JSONB,
    responsible VARCHAR(256),
    data_product_links JSONB,
    links JSONB,
    industry JSONB,
    line_of_business JSONB,
    tags JSONB,
    labels JSONB,
    documentation_labels JSONB,
    policy_level VARCHAR(256),
    custom_policy_level VARCHAR(256),
    system_instance_aware BOOLEAN,
    resource_hash VARCHAR(255),
    version_value VARCHAR(256),
    version_deprecated BOOLEAN,
    version_deprecated_since VARCHAR(256),
    version_for_removal BOOLEAN,
    ready BOOLEAN DEFAULT TRUE,
        CONSTRAINT data_products_id_ready_unique UNIQUE (id, ready),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    error JSONB
);

-- Create index for data_products table
CREATE INDEX IF NOT EXISTS data_products_app_id ON data_products(app_id);

-- Create tenants_data_products view
CREATE VIEW tenants_data_products
            (tenant_id, formation_id, id, app_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update,
                visibility, release_status, disabled, deprecation_date, sunset_date, successors, changelog_entries, type, category, entity_types, input_ports,
                output_ports, responsible, data_product_links, links, industry, line_of_business, tags, labels, documentation_labels, policy_level, custom_policy_level,
                system_instance_aware, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                d.id,
                d.app_id,
                d.ord_id,
                d.local_tenant_id,
                d.correlation_ids,
                d.title,
                d.short_description,
                d.description,
                d.package_id,
                d.last_update,
                d.visibility,
                d.release_status,
                d.disabled,
                d.deprecation_date,
                d.sunset_date,
                d.successors,
                d.changelog_entries,
                d.type,
                d.category,
                d.entity_types,
                d.input_ports,
                d.output_ports,
                d.responsible,
                d.data_product_links,
                d.links,
                d.industry,
                d.line_of_business,
                d.tags,
                d.labels,
                d.documentation_labels,
                d.policy_level,
                d.custom_policy_level,
                d.system_instance_aware,
                d.resource_hash,
                d.version_value,
                d.version_deprecated,
                d.version_deprecated_since,
                d.version_for_removal,
                d.ready,
                d.created_at,
                d.updated_at,
                d.deleted_at,
                d.error
FROM data_products d
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
           FROM apps_subaccounts) t_apps ON d.app_id = t_apps.id;

-- Create data_products_tenants view
CREATE VIEW data_products_tenants AS
SELECT d.*, ta.tenant_id, ta.owner FROM data_products AS d
                                            INNER JOIN tenant_applications ta ON ta.id = d.app_id;

-- Create data_products views for JSONB columns
-- correlation ids
CREATE VIEW correlation_ids_data_products AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.correlation_ids) AS elements;

-- successors
CREATE VIEW data_products_successors AS
SELECT id             AS data_product_id,
       elements.value AS value
FROM data_products,
     jsonb_array_elements_text(data_products.successors) AS elements;

-- changelog entries
CREATE VIEW changelog_entries_data_products AS
SELECT id                      AS data_product_id,
       entries.version         AS version,
       entries."releaseStatus" AS release_status,
       entries.date            AS date,
       entries.description     AS description,
       entries.url             AS url
FROM data_products,
     jsonb_to_recordset(data_products.changelog_entries) AS entries(version TEXT, "releaseStatus" TEXT, date TEXT, description TEXT, url TEXT);

-- entity types
CREATE VIEW entity_types_data_products AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.entity_types) AS elements;

-- input ports
CREATE VIEW input_ports_data_products AS
SELECT id                AS data_product_id,
       input_ports."ordId"       AS ord_id
FROM data_products,
     jsonb_to_recordset(data_products.input_ports) AS input_ports("ordId" TEXT);

-- output ports
CREATE VIEW output_ports_data_products AS
SELECT id                AS data_product_id,
       output_ports."ordId"       AS ord_id
FROM data_products,
     jsonb_to_recordset(data_products.output_ports) AS output_ports("ordId" TEXT);

-- data product links
CREATE VIEW data_product_links AS
SELECT id                   AS data_product_id,
       actions.type         AS type,
       actions."customType" AS custom_type,
       actions.url          AS url
FROM data_products,
     jsonb_to_recordset(data_products.data_product_links) AS actions(type TEXT, "customType" TEXT, url TEXT);

-- links
CREATE VIEW links_data_products AS
SELECT id                AS data_product_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM data_products,
     jsonb_to_recordset(data_products.links) AS links(title TEXT, description TEXT, url TEXT);

-- industry
CREATE VIEW industries_data_products AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.industry) AS elements;

-- line of business
CREATE VIEW line_of_businesses_data_products AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.line_of_business) AS elements;

-- tags
CREATE VIEW tags_data_products AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.tags) AS elements;

-- labels
CREATE VIEW ord_labels_data_products AS
SELECT id                  AS data_product_id,
       expand.key          AS key,
       elements.value      AS value
FROM data_products,
     jsonb_each(data_products.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

-- documentation labels
CREATE VIEW ord_documentation_labels_data_products AS
SELECT id                  AS data_product_id,
       expand.key          AS key,
       elements.value      AS value
FROM data_products,
     jsonb_each(data_products.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;


-- -- Drop and recreate tenants_packages view - add `runtime_restriction` column
-- DROP VIEW IF EXISTS tenants_packages;

CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, formation_id, id, ord_id, title, short_description, description, version, package_links, links,
             licence_type, tags, runtime_restriction, countries, labels, policy_level, app_id, custom_policy_level, vendor, part_of_products,
             line_of_business, industry, resource_hash, support_info)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                p.id,
                p.ord_id,
                p.title,
                p.short_description,
                p.description,
                p.version,
                p.package_links,
                p.links,
                p.licence_type,
                p.tags,
                p.runtime_restriction,
                p.countries,
                p.labels,
                p.policy_level,
                p.app_id,
                p.custom_policy_level,
                p.vendor,
                p.part_of_products,
                p.line_of_business,
                p.industry,
                p.resource_hash,
                p.support_info
FROM packages p
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                       af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON p.app_id = t_apps.id;


-- -- Drop and recreate tenants_apis view - add `responsible` and `usage` columns
-- DROP VIEW IF EXISTS tenants_apis;

CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, formation_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags,
             supported_use_cases, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible_supported, extensible_description,
             successors, resource_hash, documentation_labels, correlation_ids, direction, last_update, deprecation_date, responsible, usage)
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
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                apis.successors,
                apis.resource_hash,
                apis.documentation_labels,
                apis.correlation_ids,
                apis.direction,
                apis.last_update,
                apis.deprecation_date,
                apis.responsible,
                apis.usage
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                       af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON apis.app_id = t_apps.id,
     jsonb_to_record(apis.extensible) actions(supported text, description text);


-- -- Drop and recreate tenants_events view - add `responsible` column
-- DROP VIEW IF EXISTS tenants_events;

CREATE OR REPLACE VIEW tenants_events
            (tenant_id, formation_id, id, app_id, name, description, group_name, version_value, version_deprecated,
             version_deprecated_since, version_for_removal, ord_id, local_tenant_id, short_description,
             system_instance_aware, policy_level, custom_policy_level, changelog_entries, links, tags, countries,
             release_status, sunset_date, labels, package_id, visibility, disabled, part_of_products, line_of_business,
             industry, ready, created_at, updated_at, deleted_at, error, implementation_standard,
             custom_implementation_standard, custom_implementation_standard_description, extensible_supported,
             extensible_description, successors, resource_hash, correlation_ids, last_update, deprecation_date,
             event_resource_links, responsible)
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
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                events.successors,
                events.resource_hash,
                events.correlation_ids,
                events.last_update,
                events.deprecation_date,
                events.event_resource_links,
                events.responsible
FROM event_api_definitions events
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                       af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON events.app_id = t_apps.id,
     jsonb_to_record(events.extensible) actions(supported text, description text);


CREATE OR REPLACE VIEW entity_type_mappings_tenants(id, tenant_id, owner)
AS
SELECT DISTINCT etm.id,
                t_api_event_def.tenant_id,
                t_api_event_def.owner
FROM entity_type_mappings etm
         JOIN (SELECT a.id,
                      a.tenant_id,
                      ta.owner
               FROM tenants_apis a
                        JOIN tenant_applications ta ON ta.id = a.app_id
               UNION ALL
               SELECT e.id,
                      e.tenant_id,
                      ta.owner
               FROM tenants_events e
                        JOIN tenant_applications ta ON ta.id = e.app_id) t_api_event_def
              ON etm.api_definition_id = t_api_event_def.id OR etm.event_definition_id = t_api_event_def.id;

CREATE OR REPLACE VIEW tenants_entity_type_mappings
            (tenant_id, id, api_definition_id, event_definition_id, api_model_selectors, entity_type_targets)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                etm.id,
                etm.api_definition_id,
                etm.event_definition_id,
                etm.api_model_selectors,
                etm.entity_type_targets
FROM entity_type_mappings etm
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e) t_api_event_def
              ON etm.api_definition_id = t_api_event_def.id OR etm.event_definition_id = t_api_event_def.id;


CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format,
             event_spec_type, capability_def_id, capability_spec_type, capability_spec_format, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_capability_def.tenant_id,
                spec.id,
                spec.api_def_id,
                spec.event_def_id,
                spec.spec_data,
                spec.api_spec_format,
                spec.api_spec_type,
                spec.event_spec_format,
                spec.event_spec_type,
                spec.capability_def_id,
                spec.capability_spec_type,
                spec.capability_spec_format,
                spec.custom_type,
                spec.created_at
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e
               UNION ALL
               SELECT c.id,
                      c.tenant_id
               FROM tenants_capabilities c) t_api_event_capability_def
              ON spec.api_def_id = t_api_event_capability_def.id OR spec.event_def_id = t_api_event_capability_def.id OR
                 spec.capability_def_id = t_api_event_capability_def.id;

COMMIT;