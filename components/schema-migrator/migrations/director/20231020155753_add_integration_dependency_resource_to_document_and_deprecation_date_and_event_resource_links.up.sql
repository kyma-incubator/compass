BEGIN;

-- Drop views
DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_entity_types;

-- Alter table api_definitions - add column `deprecation_date`
ALTER TABLE api_definitions
    ADD COLUMN deprecation_date VARCHAR(256);
-- Alter table event_api_definitions - add column `deprecation_date`
ALTER TABLE event_api_definitions
    ADD COLUMN deprecation_date VARCHAR(256);
-- Alter table entity_types - add column `deprecation_date`
ALTER TABLE entity_types
    ADD COLUMN deprecation_date VARCHAR(256);

-- Recreate views for tenants_apis, tenants_events and tenants_entity_types with added `deprecation_date` column
CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, formation_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags,
             supported_use_cases, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible_supported, extensible_description, successors, resource_hash,
             documentation_labels, correlation_ids, direction, last_update, deprecation_date)
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
                apis.last_update,
                apis.deprecation_date
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
             resource_hash, correlation_ids, last_update, deprecation_date)
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
                events.last_update,
                events.deprecation_date
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
            documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal, deprecation_date)
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
                et.version_for_removal,
                et.deprecation_date
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


-- Create table for integration_dependencies
CREATE TABLE integration_dependencies
(
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id UUID NOT NULL,
        CONSTRAINT integration_dependencies_application_id_fkey FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    app_template_version_id UUID,
        CONSTRAINT integration_dependencies_app_template_version_id_fk FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
        CONSTRAINT integration_dependencies_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id),
    ord_id VARCHAR(256) NOT NULL,
        CONSTRAINT integration_dependencies_ord_id_unique UNIQUE (app_id, ord_id),
    local_tenant_id VARCHAR(256),
    correlation_ids JSONB,
    name VARCHAR(256) NOT NULL,
    short_description VARCHAR(256),
    description TEXT,
    package_id UUID,
        CONSTRAINT integration_dependencies_package_id_fk FOREIGN KEY (package_id) REFERENCES packages (id) ON DELETE CASCADE,
    last_update VARCHAR(256),
    visibility TEXT NOT NULL,
    release_status release_status NOT NULL,
    sunset_date VARCHAR(256),
    successors JSONB,
    mandatory BOOLEAN NOT NULL,
    related_integration_dependencies BOOLEAN,
    links JSONB,
    tags JSONB,
    labels JSONB,
    documentation_labels JSONB,
    resource_hash VARCHAR(255),
    version_value VARCHAR(256),
    version_deprecated BOOLEAN,
    version_deprecated_since VARCHAR(256),
    version_for_removal BOOLEAN,
    ready BOOLEAN DEFAULT TRUE,
        CONSTRAINT integration_dependency_id_ready_unique UNIQUE (id, ready),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    error JSONB
);

-- Create index for integration_dependencies table
CREATE INDEX IF NOT EXISTS integration_dependencies_app_id ON integration_dependencies (app_id);

-- Create aspects table
CREATE TABLE aspects (
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    integration_dependency_id UUID NOT NULL,
        CONSTRAINT aspects_integration_dependency_id_fkey FOREIGN KEY (integration_dependency_id) REFERENCES integration_dependencies (id) ON DELETE CASCADE,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    mandatory BOOLEAN NOT NULL,
    support_multiple_providers BOOLEAN,
    api_resources JSONB,
    event_resources JSONB,
    ready BOOLEAN DEFAULT TRUE,
        CONSTRAINT aspect_id_ready_unique UNIQUE (id, ready),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    error JSONB
);

-- Create index for aspects table
CREATE INDEX IF NOT EXISTS aspects_integration_dependency_id ON aspects (integration_dependency_id);

-- tenants_aspects view ?

-- Create views
CREATE VIEW correlation_ids_integration_dependencies AS
SELECT id                  AS integration_dependency_id,
       elements.value      AS value
FROM integration_dependencies,
     jsonb_array_elements_text(integration_dependencies.correlation_ids) AS elements;

CREATE VIEW integration_dependencies_successors AS
SELECT id             AS integration_dependency_id,
       elements.value AS value
FROM integration_dependencies,
     jsonb_array_elements_text(integration_dependencies.successors) AS elements;

-- aspects view

CREATE VIEW links_integration_dependencies AS
SELECT id                AS integration_dependency_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM integration_dependencies,
     jsonb_to_recordset(integration_dependencies.links) AS links(title TEXT, description TEXT, url TEXT);

CREATE VIEW tags_integration_dependencies AS
SELECT id                  AS integration_dependency_id,
       elements.value      AS value
FROM integration_dependencies,
     jsonb_array_elements_text(integration_dependencies.tags) AS elements;

CREATE VIEW ord_labels_integration_dependencies AS
SELECT id                  AS integration_dependency_id,
       expand.key          AS key,
       elements.value      AS value
FROM integration_dependencies,
     jsonb_each(integration_dependencies.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_integration_dependencies AS
SELECT id                  AS integration_dependency_id,
       expand.key          AS key,
       elements.value      AS value
FROM integration_dependencies,
     jsonb_each(integration_dependencies.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;


-- Create view tenants_integration_dependencies
CREATE VIEW tenants_integration_dependencies
            (tenant_id, formation_id, id, app_id, ord_id, local_tenant_id, correlation_ids, name, short_description, description, package_id,
            last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags,
            labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at,
            updated_at, deleted_at, error)
AS
SELECT DISTINCT t_apps.id,
                t_apps.formation_id,
                i.id,
                i.app_id,
                i.ord_id,
                i.local_tenant_id,
                i.correlation_ids,
                i.name,
                i.short_description,
                i.description,
                i.package_id,
                i.last_update,
                i.visibility,
                i.release_status,
                i.sunset_date,
                i.successors,
                i.mandatory,
                i.related_integration_dependencies,
                i.links,
                i.tags,
                i.labels,
                i.documentation_labels,
                i.resource_hash,
                i.version_value,
                i.version_deprecated,
                i.version_deprecated_since,
                i.version_for_removal,
                i.ready,
                i.created_at,
                i.updated_at,
                i.deleted_at,
                i.error
FROM integration_dependencies i
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
          FROM apps_subaccounts) t_apps ON i.app_id = t_apps.id;

-- Create integration_dependencies_tenants view
CREATE VIEW integration_dependencies_tenants AS
SELECT i.*, ta.tenant_id, ta.owner FROM integration_dependencies AS i
                                            INNER JOIN tenant_applications ta ON ta.id = i.app_id;


COMMIT;
