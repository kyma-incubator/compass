BEGIN;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_apps;
DROP VIEW IF EXISTS tenants_bundles;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_packages;
DROP VIEW IF EXISTS tenants_products;
DROP VIEW IF EXISTS tenants_vendors;
DROP VIEW IF EXISTS tenants_tombstones;

DROP FUNCTION IF EXISTS apps_subaccounts_func();

-----

CREATE OR REPLACE FUNCTION apps_subaccounts_func()
    returns TABLE(id uuid, tenant_id text)
    language plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT l.app_id                 AS id,
               asa.target_tenant_id::text AS tenant_id
        FROM labels l
                 -- 2) Get subaccounts in those scenarios (Putting a subaccount in a
                 -- scenario will reflect on creating an ASA for the subaccount.
                 JOIN automatic_scenario_assignments asa
                      ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario::text
             -- 1) Get all scenario labels for applications
        WHERE l.app_id IS NOT NULL
          AND l.key::text = 'scenarios'::text;
END
$$;

-----

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
                CASE
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'ODATA'::text THEN 'odata-v2'::text
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'OPEN_API'::text THEN 'rest'::text
                    ELSE apis.api_protocol::text
                    END AS api_protocol,
                apis.tags,
                apis.countries,
                apis.links,
                apis.api_resource_links,
                apis.release_status,
                apis.sunset_date,
                apis.changelog_entries,
                apis.labels,
                apis.package_id,
                apis.visibility::text as visibility,
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
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON apis.app_id = t_apps.id
         LEFT JOIN specifications specs ON apis.id = specs.api_def_id;

-----

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, product_type)
AS
SELECT DISTINCT t_apps.tenant_id,
                apps.id,
                apps.name,
                apps.description,
                apps.status_condition,
                apps.status_timestamp,
                apps.healthcheck_url,
                apps.integration_system_id,
                apps.provider_name,
                apps.base_url,
                apps.labels,
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                tmpl.name AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON apps.id = t_apps.id;

-----

CREATE OR REPLACE VIEW tenants_bundles
            (tenant_id, id, app_id, name, description, instance_auth_request_json_schema,
             default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready,
             created_at, updated_at, deleted_at, error, correlation_ids)
AS
SELECT DISTINCT t_apps.tenant_id,
                b.id,
                b.app_id,
                b.name,
                b.description,
                b.instance_auth_request_json_schema,
                b.default_instance_auth,
                b.ord_id,
                b.short_description,
                b.links,
                b.labels,
                b.credential_exchange_strategies,
                b.ready,
                b.created_at,
                b.updated_at,
                b.deleted_at,
                b.error,
                b.correlation_ids
FROM bundles b
         JOIN (SELECT a1.id,
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON b.app_id = t_apps.id;

-----

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
                events.visibility::text as visibility,
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
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON events.app_id = t_apps.id;

-----

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
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON p.app_id = t_apps.id;

-----

CREATE OR REPLACE VIEW tenants_products
            (tenant_id, ord_id, app_id, title, short_description, vendor, parent, labels,
             correlation_ids, id, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
                p.ord_id,
                p.app_id,
                p.title,
                p.short_description,
                p.vendor,
                p.parent,
                p.labels,
                p.correlation_ids,
                p.id,
                p.documentation_labels
FROM products p
         JOIN (SELECT a1.id,
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON p.app_id = t_apps.id OR p.app_id IS NULL;

-----

CREATE OR REPLACE VIEW tenants_vendors
            (tenant_id, ord_id, app_id, title, labels, partners, id, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
                v.ord_id,
                v.app_id,
                v.title,
                v.labels,
                v.partners,
                v.id,
                v.documentation_labels
FROM vendors v
         JOIN (SELECT a1.id,
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON v.app_id = t_apps.id OR v.app_id IS NULL;

-----

CREATE OR REPLACE VIEW tenants_tombstones(tenant_id, ord_id, app_id, removal_date, id)
AS
SELECT DISTINCT t_apps.tenant_id,
                t.ord_id,
                t.app_id,
                t.removal_date,
                t.id
FROM tombstones t
         JOIN (SELECT a1.id,
                      a1.tenant_id::text AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) t_apps
              ON t.app_id = t_apps.id;

-----

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

COMMIT;
