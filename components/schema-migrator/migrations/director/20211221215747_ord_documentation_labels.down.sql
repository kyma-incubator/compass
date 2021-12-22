BEGIN;

-- ord labels view
DROP VIEW IF EXISTS ord_documentation_labels;
-- packages
ALTER TABLE packages DROP COLUMN documentation_labels;

-- consumption bundles
ALTER TABLE bundles DROP COLUMN documentation_labels;

-- api resource
DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
ALTER TABLE api_definitions DROP COLUMN documentation_labels;

CREATE VIEW tenants_apis AS
SELECT DISTINCT t_apps.tenant_id AS tenant_id,
                t_apps.provider_tenant_id  AS provider_tenant_id,
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
                apis.resource_hash
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id, a_s.tenant_id, (SELECT id::text FROM business_tenant_mappings WHERE external_tenant = cpr.provider_tenant) AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? (SELECT external_tenant FROM business_tenant_mappings WHERE id = a_s.tenant_id::uuid)) t_apps ON apis.app_id = t_apps.id
         LEFT JOIN specifications specs ON apis.id = specs.api_def_id;

CREATE OR REPLACE VIEW tenants_specifications AS
SELECT DISTINCT t_api_event_def.tenant_id AS tenant_id,
                t_api_event_def.provider_tenant_id  AS provider_tenant_id,
                spec.*
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id::text,
                      a.provider_tenant_id::text
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id::text,
                      e.provider_tenant_id::text
               FROM tenants_events e) t_api_event_def
              ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;

-- event resource
ALTER TABLE event_api_definitions DROP COLUMN documentation_labels;

-- product
ALTER TABLE products DROP COLUMN documentation_labels;

-- vendor
ALTER TABLE vendors DROP COLUMN documentation_labels;

-- system instance
ALTER TABLE applications DROP COLUMN documentation_labels;

COMMIT;
