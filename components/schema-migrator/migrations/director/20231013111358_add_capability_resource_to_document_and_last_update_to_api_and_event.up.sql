BEGIN;

-- Drop views
DROP VIEW IF EXISTS api_specifications_tenants;
DROP VIEW IF EXISTS event_specifications_tenants;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;

-- Alter tables - add `last_update` to API and Event
ALTER TABLE api_definitions
    ADD COLUMN last_update VARCHAR(256);

ALTER TABLE event_api_definitions
    ADD COLUMN last_update VARCHAR(256);


-- Recreate views for tenant_apis and tenant_events with added `last_update` column
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


-- Create db type for Capability Type
CREATE TYPE capability_type AS ENUM ('custom', 'sap.mdo:mdi-capability:v1');

-- Create table for capabilities
CREATE TABLE IF NOT EXISTS capabilities
(
    id                      UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id                  UUID NOT NULL,
        CONSTRAINT capabilities_application_id_fkey FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    ord_id                  VARCHAR(256) NOT NULL,
        CONSTRAINT capabilities_ord_id_unique UNIQUE (app_id, ord_id),
    app_template_version_id UUID,
        CONSTRAINT capabilities_app_template_version_id_fk FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
        CONSTRAINT capabilities_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id),
    package_id              UUID,
        CONSTRAINT capabilities_package_id_fk FOREIGN KEY (package_id) REFERENCES packages (id) ON DELETE CASCADE,
    name                    VARCHAR(256) NOT NULL,
    description             TEXT,
    type                    capability_type NOT NULL,
    custom_type VARCHAR(256),
    local_tenant_id VARCHAR(256),
    short_description VARCHAR(256),
    system_instance_aware BOOLEAN,
    tags JSONB,
    links JSONB,
    release_status release_status NOT NULL,
    labels JSONB,
    visibility TEXT NOT NULL,
    resource_hash VARCHAR(255),
    documentation_labels JSONB,
    correlation_ids JSONB,
    version_value VARCHAR(256),
    version_deprecated BOOLEAN,
    version_deprecated_since VARCHAR(256),
    version_for_removal BOOLEAN,
    ready BOOLEAN DEFAULT TRUE,
        CONSTRAINT capability_id_ready_unique UNIQUE (id, ready),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    error JSONB,
    last_update VARCHAR(256)
);

-- Create db type for Capability Specification type
CREATE TYPE capability_spec_type AS ENUM ('custom', 'sap.mdo:mdi-capability-definition:v1');

-- Create db type for Capability Specification format
CREATE TYPE capability_spec_format AS ENUM (
    'application/json',
    'text/yaml',
    'application/xml',
    'text/plain',
    'application/octet-stream'
);

-- Create index for capabilities table
CREATE INDEX IF NOT EXISTS capabilities_app_id ON capabilities (app_id);

-- Alter table specifications - add capability_def_id, capability_spec_type and capability_spec_format
ALTER TABLE specifications
    ADD COLUMN capability_def_id UUID,
    ADD CONSTRAINT specifications_capability_id_fkey FOREIGN KEY (capability_def_id) REFERENCES capabilities (id) ON DELETE CASCADE,
    ADD COLUMN capability_spec_type capability_spec_type,
    ADD COLUMN capability_spec_format capability_spec_format;

-- Create index for specifications table
CREATE INDEX capability_specifications_tenants_app_id ON specifications (capability_def_id);

-- Create helper views for tags, links, ord_labels, correlation_ids and ord_documentation_labels of Capability
CREATE VIEW tags_capabilities AS
SELECT id                  AS capability_id,
       elements.value      AS value
FROM capabilities,
     jsonb_array_elements_text(capabilities.tags) AS elements;

CREATE VIEW links_capabilities AS
SELECT id                AS capability_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM capabilities,
     jsonb_to_recordset(capabilities.links) AS links(title TEXT, description TEXT, url TEXT);

CREATE VIEW ord_labels_capabilities AS
SELECT id                  AS capability_id,
       expand.key          AS key,
       elements.value      AS value
FROM capabilities,
     jsonb_each(capabilities.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW correlation_ids_capabilities AS
SELECT id                  AS capability_id,
       elements.value      AS value
FROM capabilities,
     jsonb_array_elements_text(capabilities.correlation_ids) AS elements;

CREATE VIEW ord_documentation_labels_capabilities AS
SELECT id                  AS capability_id,
       expand.key          AS key,
       elements.value      AS value
FROM capabilities,
     jsonb_each(capabilities.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

-- Create views tenants_capabilities, capability_definitions, capability_specifications_tenants, capabilities_tenants and capability_specifications_fetch_requests_tenants
-- Recreate views api_specifications_tenants and event_specifications_tenants
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

CREATE VIEW capability_definitions AS
SELECT capability_def_id,
       capability_spec_type::text AS type,
        custom_type,
       format('/capability/%s/specification/%s', capability_def_id, id) AS url,
       capability_spec_format::text AS media_type
FROM specifications
WHERE capability_def_id IS NOT NULL;

CREATE OR REPLACE VIEW api_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                             INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                             INNER JOIN tenant_applications ta ON ta.id = ad.app_id);


CREATE OR REPLACE VIEW event_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                            INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                            INNER JOIN tenant_applications ta ON ta.id = ead.app_id);

CREATE VIEW capability_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                             INNER JOIN capabilities AS cd ON cd.id = s.capability_def_id
                                             INNER JOIN tenant_applications ta ON ta.id = cd.app_id);


CREATE VIEW capability_specifications_fetch_requests_tenants AS
(SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                              INNER JOIN specifications s ON fr.spec_id = s.id
                                              INNER JOIN capabilities AS cd ON cd.id = s.capability_def_id
                                              INNER JOIN tenant_applications ta ON ta.id = cd.app_id);


CREATE VIEW capabilities_tenants AS
SELECT cd.*, ta.tenant_id, ta.owner FROM capabilities AS cd
                                             INNER JOIN tenant_applications ta ON ta.id = cd.app_id;

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
