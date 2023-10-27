BEGIN;

-- Drop views
DROP VIEW IF EXISTS aspect_event_resources_subset;
DROP VIEW IF EXISTS aspect_event_resources;
DROP VIEW IF EXISTS aspect_api_resources;
DROP VIEW IF EXISTS aspects_tenants;
DROP VIEW IF EXISTS tenants_aspects;

DROP VIEW IF EXISTS ord_documentation_labels_integration_dependencies;
DROP VIEW IF EXISTS ord_labels_integration_dependencies;
DROP VIEW IF EXISTS tags_integration_dependencies;
DROP VIEW IF EXISTS links_integration_dependencies;
DROP VIEW IF EXISTS tenants_integration_dependencies;
DROP VIEW IF EXISTS related_integration_dependencies;
DROP VIEW IF EXISTS integration_dependencies_successors;
DROP VIEW IF EXISTS correlation_ids_integration_dependencies;
DROP VIEW IF EXISTS integration_dependencies_tenants;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_entity_types;

-- Drop index for event_api_definitions table on ord_id column
DROP INDEX IF EXISTS event_api_def_ord_id;
-- Drop index for aspects table
DROP INDEX IF EXISTS aspects_app_id;
-- Drop index for integration_dependencies table
DROP INDEX IF EXISTS integration_dependencies_app_id;

-- Drop table aspects
DROP TABLE IF EXISTS aspects;
-- Drop table integration_dependencies
DROP TABLE IF EXISTS integration_dependencies;

-- Alter tables - remove `deprecation_date` from API, Event and Entity Type
ALTER TABLE api_definitions
    DROP COLUMN deprecation_date;

ALTER TABLE event_api_definitions
    DROP COLUMN deprecation_date;

ALTER TABLE entity_types
    DROP COLUMN deprecation_date;

-- Recreate views for tenant_apis, tenant_events and tenant_entity_types with removed `deprecation_date` column
CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, formation_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags,
             supported_use_cases, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible_supported, extensible_description, successors, resource_hash,
             documentation_labels, correlation_ids, direction, last_update)
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
                actions.supported,
                actions.description,
                apis.successors,
                apis.resource_hash,
                apis.documentation_labels,
                apis.correlation_ids,
                apis.direction,
                apis.last_update
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
             system_instance_aware, policy_level, custom_policy_level, changelog_entries, links, tags,
             countries, release_status, sunset_date, labels, package_id, visibility, disabled, part_of_products,
             line_of_business, industry, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, extensible_supported, extensible_description, successors,
             resource_hash, correlation_ids, last_update)
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
                actions.supported,
                actions.description,
                events.successors,
                events.resource_hash,
                events.correlation_ids,
                events.last_update
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


CREATE OR REPLACE VIEW tenants_entity_types
            (tenant_id, formation_id, id, ord_id, app_id, local_id, level, title, short_description, description, system_instance_aware,
            changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level,
            custom_policy_level, release_status, sunset_date, successors, extensible_supported, extensible_description, tags, labels,
            documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                et.id,
                et.ord_id,
                et.app_id,
                et.local_id,
                et.level,
                et.title,
                et.short_description,
                et.description,
                et.system_instance_aware,
                et.changelog_entries,
                et.package_id,
                et.visibility,
                et.links,
                et.part_of_products,
                et.last_update,
                et.policy_level,
                et.custom_policy_level,
                et.release_status,
                et.sunset_date,
                et.successors,
                actions.supported,
                actions.description,
                et.tags,
                et.labels,
                et.documentation_labels,
                et.resource_hash,
                et.version_value,
                et.version_deprecated,
                et.version_deprecated_since,
                et.version_for_removal
FROM entity_types et
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
              ON et.app_id = t_apps.id,
     jsonb_to_record(et.extensible) actions(supported text, description text);

-- Recreate view tenants_specifications
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
              ON spec.api_def_id = t_api_event_capability_def.id OR spec.event_def_id = t_api_event_capability_def.id or spec.capability_def_id = t_api_event_capability_def.id;

COMMIT;
