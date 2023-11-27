BEGIN;

DROP VIEW IF EXISTS tenants_specifications; -- this one won't be changed but it uses tenants_apis/events and it has to be dropped so the other two can be as well
DROP VIEW IF EXISTS tenants_entity_type_mappings; -- this one won't be changed but it uses tenants_apis/events and it has to be dropped so the other two can be as well
DROP VIEW IF EXISTS entity_type_mappings_tenants; -- this one won't be changed but it uses tenants_apis/events and it has to be dropped so the other two can be as well
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_packages;


-- the new filtering only by formation_id (with default tenant_id) was missed in the initial migration, so it is adapted here

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
