BEGIN;

UPDATE api_definitions
SET visibility = 'public'
WHERE visibility IS NULL;

ALTER TABLE api_definitions
ALTER COLUMN visibility SET NOT NULL;

UPDATE event_api_definitions
SET visibility = 'public'
WHERE visibility IS NULL;

ALTER TABLE event_api_definitions
ALTER COLUMN visibility SET NOT NULL;

--- changes to tenants_apis and tenants_events ORD views

DROP VIEW IF EXISTS tenants_specifications; -- drop tenant_specifications because tenant_apis and tenant_events depends on it. There are no changes on this view.
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_events;

-- cast visibility field to text
CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, provider_tenant_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, short_description,
             system_instance_aware, api_protocol, tags, countries, links, api_resource_links, release_status,
             sunset_date, changelog_entries, labels, package_id, visibility, disabled, part_of_products,
             line_of_business, industry, ready, created_at, updated_at, deleted_at, error, implementation_standard,
             custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible,
             successors, resource_hash, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                      a1.tenant_id::text AS tenant_id,
                      a1.tenant_id::text AS provider_tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text,
                      apps_subaccounts_func.provider_tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id, provider_tenant_id)
               UNION ALL
               SELECT ta.id AS app_id, ta.tenant_id::text AS consumer_tenant, tenant_runtimes.tenant_id::text AS provider_tenant
               FROM (SELECT labels.runtime_id, v ->> 0 AS consumer_tenant
                     FROM labels
                              JOIN jsonb_array_elements(labels.value) AS v ON TRUE
                     WHERE key = 'consumer_subaccount_ids') AS t_rts -- Get runtime and external consumer IDs pairs
                        JOIN business_tenant_mappings t ON t_rts.consumer_tenant = t.external_tenant -- Get runtime and internal consumer IDs pairs
                        JOIN apps_subaccounts_func() ta ON t.id = ta.tenant_id -- Get applications for consumer tenants
                        JOIN tenant_runtimes ON t_rts.runtime_id = tenant_runtimes.id) t_apps
              ON apis.app_id = t_apps.id
         LEFT JOIN specifications specs ON apis.id = specs.api_def_id;

-----

-- cast visibility field to text
CREATE OR REPLACE VIEW tenants_events
            (tenant_id, provider_tenant_id, id, app_id, name, description, group_name, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, short_description,
             system_instance_aware, changelog_entries, links, tags, countries, release_status, sunset_date, labels,
             package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, extensible, successors, resource_hash)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                      a1.tenant_id::text AS tenant_id,
                      a1.tenant_id::text AS provider_tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text,
                      apps_subaccounts_func.provider_tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id, provider_tenant_id)
               UNION ALL
               SELECT ta.id AS app_id, ta.tenant_id::text AS consumer_tenant, tenant_runtimes.tenant_id::text AS provider_tenant
               FROM (SELECT labels.runtime_id, v ->> 0 AS consumer_tenant
                     FROM labels
                              JOIN jsonb_array_elements(labels.value) AS v ON TRUE
                     WHERE key = 'consumer_subaccount_ids') AS t_rts -- Get runtime and external consumer IDs pairs
                        JOIN business_tenant_mappings t ON t_rts.consumer_tenant = t.external_tenant -- Get runtime and internal consumer IDs pairs
                        JOIN apps_subaccounts_func() ta ON t.id = ta.tenant_id -- Get applications for consumer tenants
                        JOIN tenant_runtimes ON t_rts.runtime_id = tenant_runtimes.id) t_apps
              ON events.app_id = t_apps.id;

-----

CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, provider_tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type,
             event_spec_format, event_spec_type, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                t_api_event_def.provider_tenant_id,
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
                      a.tenant_id,
                      a.provider_tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id,
                      e.provider_tenant_id
               FROM tenants_events e) t_api_event_def
              ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;

COMMIT;
