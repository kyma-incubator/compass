
BEGIN;

-- We need to drop these views because they references the 'policy_level' type. After we modify the `policy_level` type we can recreate the views as is done eventually in this script.
DROP VIEW IF EXISTS tenants_packages;
DROP VIEW IF EXISTS packages_tenants;


DROP VIEW IF EXISTS tenants_specifications; -- drop tenant_specifications because tenants_apis and tenants_events depend on it. There are no specific changes on this view.
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;

ALTER TYPE policy_level RENAME TO policy_level_old;

CREATE TYPE policy_level AS ENUM ('custom', 'sap:core:v1','sap:partner:v1', 'none');

ALTER TABLE packages
    ALTER COLUMN policy_level TYPE policy_level
        USING policy_level::text::policy_level;

ALTER TABLE api_definitions
    DROP COLUMN policy_level,
    DROP COLUMN custom_policy_level;

ALTER TABLE event_api_definitions
    DROP COLUMN policy_level,
    DROP COLUMN custom_policy_level;

CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, short_description,
             system_instance_aware, api_protocol, tags, countries, links, api_resource_links, release_status,
             sunset_date, changelog_entries, labels, package_id, visibility, disabled, part_of_products,
             line_of_business, industry, ready, created_at, updated_at, deleted_at, error, implementation_standard,
             custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible,
             successors, resource_hash, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
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
                apis.short_description,
                apis.system_instance_aware,
                apis.api_protocol,
                apis.tags,
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
                apis.extensible,
                apis.successors,
                apis.resource_hash,
                apis.documentation_labels
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id
               FROM apps_subaccounts) t_apps
              ON apis.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_events
            (tenant_id, id, app_id, name, description, group_name, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, short_description,
             system_instance_aware, changelog_entries, links, tags, countries, release_status, sunset_date, labels,
             package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, extensible, successors, resource_hash)
AS
SELECT DISTINCT t_apps.tenant_id,
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
                events.short_description,
                events.system_instance_aware,
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
                events.extensible,
                events.successors,
                events.resource_hash
FROM event_api_definitions events
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id
               FROM apps_subaccounts) t_apps
              ON events.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type,
             event_spec_format, event_spec_type, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                spec.id,
                spec.api_def_id,
                spec.event_def_id,
                spec.spec_data,
                spec.api_spec_format,
                spec.api_spec_type,
                spec.event_spec_format,
                spec.event_spec_type,
                spec.custom_type,
                spec.created_at
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e) t_api_event_def
              ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;

CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, id, ord_id, title, short_description, description, version, package_links,
             links, licence_type, tags, countries, labels, policy_level, app_id, custom_policy_level, vendor,
             part_of_products, line_of_business, industry, resource_hash, support_info)
AS
SELECT DISTINCT t_apps.tenant_id,
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
                      a1.tenant_id AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id
               FROM apps_subaccounts) t_apps
              ON p.app_id = t_apps.id;

CREATE OR REPLACE VIEW packages_tenants
            (id, ord_id, title, short_description, description, version, package_links, links, licence_type, tags,
             countries, labels, policy_level, app_id, custom_policy_level, vendor, part_of_products, line_of_business,
             industry, resource_hash, documentation_labels, support_info, tenant_id, owner)
AS
SELECT p.id,
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
       p.documentation_labels,
       p.support_info,
       ta.tenant_id,
       ta.owner
FROM packages p
         JOIN tenant_applications ta ON ta.id = p.app_id;

DROP TYPE policy_level_old;

COMMIT;