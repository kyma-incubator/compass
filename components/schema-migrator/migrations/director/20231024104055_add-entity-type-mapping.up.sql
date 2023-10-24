BEGIN;

ALTER TABLE capabilities ADD COLUMN related_entity_types JSONB;


DROP VIEW IF EXISTS entity_types_capabilities;

CREATE OR REPLACE VIEW entity_types_capabilities AS
SELECT id                  AS capability_id,
       elements.value      AS value
FROM capabilities,
     jsonb_array_elements_text(capabilities.related_entity_types) AS elements;


DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_capabilities;

CREATE OR REPLACE VIEW tenants_capabilities
            (tenant_id, formation_id, id, app_id, name, description, type, custom_type, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, tags, related_entity_types, links, release_status,
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
                c.related_entity_types,
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


COMMIT;
