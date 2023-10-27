BEGIN;

DROP VIEW IF EXISTS tenants_entity_types;
ALTER TABLE entity_types RENAME COLUMN  local_tenant_id to local_id;

DROP VIEW IF EXISTS tenants_entity_types;

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


DROP VIEW IF EXISTS tenants_entity_type_mappings;
DROP VIEW IF EXISTS entity_type_mappings_tenants;

DROP VIEW IF EXISTS entity_type_targets_entity_type_mappings;
DROP VIEW IF EXISTS api_model_selectors_entity_type_mappings;

DROP INDEX IF EXISTS entity_type_mappings_api_definition_id;
DROP INDEX IF EXISTS entity_type_mappings_event_definition_id;
DROP TABLE entity_type_mappings;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_capabilities;

CREATE OR REPLACE VIEW tenants_capabilities
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

DROP VIEW IF EXISTS entity_types_capabilities;

ALTER TABLE capabilities DROP related_entity_types;

COMMIT;
