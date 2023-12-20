BEGIN;

-- Drop views
DROP VIEW IF EXISTS ord_documentation_labels_data_products;
DROP VIEW IF EXISTS ord_labels_data_products;
DROP VIEW IF EXISTS tags_data_products;
DROP VIEW IF EXISTS line_of_businesses_data_products;
DROP VIEW IF EXISTS industries_data_products;
DROP VIEW IF EXISTS links_data_products;
DROP VIEW IF EXISTS data_product_links;
DROP VIEW IF EXISTS output_ports_data_products;
DROP VIEW IF EXISTS input_ports_data_products;
DROP VIEW IF EXISTS entity_types_data_products;
DROP VIEW IF EXISTS changelog_entries_data_products;
DROP VIEW IF EXISTS data_products_successors;
DROP VIEW IF EXISTS correlation_ids_data_products;

DROP VIEW IF EXISTS tenants_data_products;
DROP VIEW IF EXISTS data_products_tenants;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_entity_type_mappings;
DROP VIEW IF EXISTS entity_type_mappings_tenants;

-- Drop and recreate tenants_apis view - remove `responsible` and `usage` columns
DROP VIEW IF EXISTS tenants_apis;
-- Drop and recreate tenants_events view - remove `responsible` column
DROP VIEW IF EXISTS tenants_events;
-- Drop and recreate tenants_packages view - remove `runtime_restriction` column
DROP VIEW IF EXISTS tenants_packages;


-- Drop index for data_products table
DROP INDEX IF EXISTS data_products_app_id;

-- Drop data_products table
DROP TABLE IF EXISTS data_products;

-- Drop `runtimeRestriction` column from Package
ALTER TABLE packages
    DROP COLUMN runtime_restriction;

-- Drop `responsible` and `usage` column from API
ALTER TABLE api_definitions
    DROP COLUMN responsible,
    DROP COLUMN usage;

-- Drop `responsible` column from Event
ALTER TABLE event_api_definitions
    DROP COLUMN responsible;


CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, formation_id, id, ord_id, title, short_description, description, version, package_links, links,
             licence_type, tags, countries, labels, policy_level, app_id, custom_policy_level, vendor, part_of_products,
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


CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, formation_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags,
             supported_use_cases, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible_supported, extensible_description,
             successors, resource_hash, documentation_labels, correlation_ids, direction, last_update, deprecation_date)
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
                apis.deprecation_date
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


CREATE OR REPLACE VIEW tenants_events
            (tenant_id, formation_id, id, app_id, name, description, group_name, version_value, version_deprecated,
             version_deprecated_since, version_for_removal, ord_id, local_tenant_id, short_description,
             system_instance_aware, policy_level, custom_policy_level, changelog_entries, links, tags, countries,
             release_status, sunset_date, labels, package_id, visibility, disabled, part_of_products, line_of_business,
             industry, ready, created_at, updated_at, deleted_at, error, implementation_standard,
             custom_implementation_standard, custom_implementation_standard_description, extensible_supported,
             extensible_description, successors, resource_hash, correlation_ids, last_update, deprecation_date,
             event_resource_links)
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
                events.event_resource_links
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