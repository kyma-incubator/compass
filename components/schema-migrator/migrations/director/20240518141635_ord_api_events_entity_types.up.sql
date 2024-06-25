BEGIN;

-- create views that will represent the relation between events/apis and entityTypes
-- the joins are based on the ordId of the entityTypeMapping.entityTypeTargets.ordId and the entityType.ordId
-- then the event/api internal id is taken from the entityTypeMapping

CREATE OR REPLACE VIEW tenants_event_entity_types
            (event_definition_id, tenant_id, formation_id, id, ord_id, app_id, local_tenant_id, level, title, short_description, description,
             system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update,
             policy_level, custom_policy_level, release_status, sunset_date, successors, extensible_supported,
             extensible_description, tags, labels, documentation_labels, resource_hash, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, deprecation_date)
AS
SELECT DISTINCT etm.event_definition_id,
                t_apps.tenant_id,
                t_apps.formation_id,
                et.id,
                et.ord_id,
                et.app_id,
                et.local_tenant_id,
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
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                et.tags,
                et.labels,
                et.documentation_labels,
                et.resource_hash,
                et.version_value,
                et.version_deprecated,
                et.version_deprecated_since,
                et.version_for_removal,
                et.deprecation_date
FROM entity_types et
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
               FROM apps_subaccounts) t_apps ON et.app_id = t_apps.id, entity_type_mappings etm, jsonb_to_recordset(etm.entity_type_targets) AS entries("ordId" TEXT, "correlationId" TEXT), JSONB_TO_RECORD(et.extensible) actions(supported text, description text) WHERE entries."ordId" = et.ord_id AND etm.event_definition_id IS NOT NULL;


CREATE OR REPLACE VIEW tenants_api_entity_types
            (api_definition_id, tenant_id, formation_id, id, ord_id, app_id, local_tenant_id, level, title, short_description, description,
             system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update,
             policy_level, custom_policy_level, release_status, sunset_date, successors, extensible_supported,
             extensible_description, tags, labels, documentation_labels, resource_hash, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, deprecation_date)
AS
SELECT DISTINCT etm.api_definition_id,
                t_apps.tenant_id,
                t_apps.formation_id,
                et.id,
                et.ord_id,
                et.app_id,
                et.local_tenant_id,
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
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                et.tags,
                et.labels,
                et.documentation_labels,
                et.resource_hash,
                et.version_value,
                et.version_deprecated,
                et.version_deprecated_since,
                et.version_for_removal,
                et.deprecation_date
FROM entity_types et
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
               FROM apps_subaccounts) t_apps ON et.app_id = t_apps.id, entity_type_mappings etm, jsonb_to_recordset(etm.entity_type_targets) AS entries("ordId" TEXT, "correlationId" TEXT), JSONB_TO_RECORD(et.extensible) actions(supported text, description text) WHERE entries."ordId" = et.ord_id AND etm.api_definition_id IS NOT NULL;

COMMIT;
