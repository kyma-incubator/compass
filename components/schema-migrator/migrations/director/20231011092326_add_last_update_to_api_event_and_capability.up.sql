BEGIN;

-- Drop views
DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_capabilities;

-- Alter tables - add `last_update` to API, Event and Capability
ALTER TABLE api_definitions
    ADD COLUMN last_update VARCHAR(256);

ALTER TABLE event_api_definitions
    ADD COLUMN last_update VARCHAR(256);

ALTER TABLE capabilities
    ADD COLUMN last_update VARCHAR(256);

-- Recreate views for tenant_apis, tenant_events and tenant_capabilities with added `last_update` column
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

CREATE VIEW tenants_capabilities
            (tenant_id, formation_id, id, app_id, name, description, type, custom_type, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, tags, links, release_status,
             labels, package_id, visibility, ready, created_at,
             updated_at, deleted_at, error, resource_hash,
             documentation_labels, correlation_ids, last_update)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                c.id,
                c.app_id,
                c.name,
                c.description,
                c.type,
                c.custom_type,
                c.version_value,
                c.version_deprecated,
                c.version_deprecated_since,
                c.version_for_removal,
                c.ord_id,
                c.local_tenant_id,
                c.short_description,
                c.system_instance_aware,
                c.tags,
                c.links,
                c.release_status,
                c.labels,
                c.package_id,
                c.visibility,
                c.ready,
                c.created_at,
                c.updated_at,
                c.deleted_at,
                c.error,
                c.resource_hash,
                c.documentation_labels,
                c.correlation_ids,
                c.last_update
FROM capabilities c
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
               FROM apps_subaccounts) t_apps ON c.app_id = t_apps.id;


-- Recreate view for tenants_specifications
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
