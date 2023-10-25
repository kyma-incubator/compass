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


CREATE TABLE entity_type_mappings
(
    id                          UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    ready                       BOOLEAN DEFAULT TRUE,
        CONSTRAINT entitytypemaping_id_ready_unique UNIQUE (id, ready),
    api_definition_id             UUID,
        CONSTRAINT entity_type_maping_api_definition_id_fk FOREIGN KEY (api_definition_id) REFERENCES api_definitions (id) ON DELETE CASCADE,
    event_definition_id           UUID,
        CONSTRAINT entity_type_maping_event_definition_id_fk FOREIGN KEY (event_definition_id) REFERENCES event_api_definitions (id) ON DELETE CASCADE,
    created_at                  TIMESTAMP,
    updated_at                  TIMESTAMP,
    deleted_at                  TIMESTAMP,
    error                       JSONB,
    api_model_selectors         JSONB,
    entity_type_targets         JSONB
);

CREATE INDEX entity_type_mappings_api_definition_id ON entity_type_mappings (api_definition_id);
CREATE INDEX entity_type_mappings_event_definition_id ON entity_type_mappings (event_definition_id);

DROP VIEW IF EXISTS api_model_selectors_entity_type_mappings;

CREATE VIEW api_model_selectors_entity_type_mappings AS
SELECT id                      AS entity_type_mapping_id,
       entries.type            AS type,
       entries."entitySetName" AS entity_set_name,
       entries."jsonPointer"   AS json_pointer
FROM entity_type_mappings,
     jsonb_to_recordset(entity_type_mappings.api_model_selectors) AS entries(type TEXT, "entitySetName" TEXT, "jsonPointer" TEXT);

DROP VIEW IF EXISTS entity_type_targets_entity_type_mappings;

CREATE VIEW entity_type_targets_entity_type_mappings AS
SELECT id                      AS entity_type_mapping_id,
       entries."ordId"         AS ord_id,
       entries."correlationId" AS correlation_id
FROM entity_type_mappings,
     jsonb_to_recordset(entity_type_mappings.entity_type_targets) AS entries("ordId" TEXT, "correlationId" TEXT);

DROP VIEW IF EXISTS tenants_entity_type_mappings;

CREATE OR REPLACE VIEW tenants_entity_type_mappings
            (tenant_id, formation_id, id, api_definition_id, event_definition_id)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                etm.id,
                etm.api_definition_id,
                etm.event_definition_id
FROM entity_type_mappings etm
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
               FROM apps_subaccounts) t_apps, 
              ON et.app_id = t_apps.id
// TODO - tenants_entity_type_mappings not finished


DROP VIEW IF EXISTS tenants_entity_types;

ALTER TABLE entity_types RENAME COLUMN  local_id to local_tenant_id;

CREATE OR REPLACE VIEW tenants_entity_types
            (tenant_id, formation_id, id, ord_id, app_id, local_tenant_id, level, title, short_description, description, system_instance_aware, 
            changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level,
            custom_policy_level, release_status, sunset_date, successors, extensible_supported, extensible_description, tags, labels, 
            documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal)
AS
SELECT DISTINCT t_apps.tenant_id,
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

COMMIT;
