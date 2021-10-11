BEGIN;

DROP VIEW tenants_apps;
DROP VIEW tenants_bundles;
DROP VIEW tenants_specifications;
DROP VIEW tenants_apis;
DROP VIEW tenants_events;

CREATE OR REPLACE VIEW tenants_apps  AS
SELECT DISTINCT t_apps.tenant_id, apps.id, apps.name, apps.description, apps.status_condition,
                apps.status_timestamp, apps.healthcheck_url, apps.integration_system_id,
                apps.provider_name, apps.base_url, apps.labels, apps.ready, apps.created_at,
                apps.updated_at, apps.deleted_at, apps.error, apps.app_template_id, apps.correlation_ids,
                apps.system_number, tmpl.name as product_type
FROM applications AS apps
         LEFT JOIN app_templates AS tmpl ON apps.app_template_id = tmpl.id
         INNER JOIN (
--  select GAs
    SELECT a1.id, a1.tenant_id::text FROM applications AS a1
    UNION ALL
--  select CRM
    SELECT a.id, t.parent::text as tenant_id FROM applications AS a
                                                INNER JOIN  business_tenant_mappings AS t ON t.id = a.tenant_id WHERE t.parent IS NOT NULL
    UNION ALL
--  select SA
    SELECT l.app_id as id, asa.selector_value::text as tenant_id FROM labels AS l
                                                                          INNER JOIN automatic_scenario_assignments AS asa ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario AND asa.selector_key='global_subaccount_id'
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON apps.id = t_apps.id;

CREATE OR REPLACE VIEW tenants_apis  AS
SELECT DISTINCT t_apps.tenant_id, apis.id, apis.app_id, apis.name, apis.description, apis.group_name, apis.default_auth,
                apis.version_value, apis.version_deprecated, apis.version_deprecated_since, apis.version_for_removal,
                apis.ord_id, apis.short_description, apis.system_instance_aware,
                CASE
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'ODATA' THEN 'odata-v2'::text
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'OPEN_API' THEN 'rest'::text
                    ELSE apis.api_protocol::text
                    END                                            AS api_protocol,
                apis.tags, apis.countries, apis.links, apis.api_resource_links, apis.release_status, apis.sunset_date,
                apis.changelog_entries, apis.labels, apis.package_id, apis.visibility, apis.disabled, apis.part_of_products,
                apis.line_of_business, apis.industry, apis.ready, apis.created_at, apis.updated_at, apis.deleted_at, apis.error,
                apis.implementation_standard, apis.custom_implementation_standard, apis.custom_implementation_standard_description,
                apis.target_urls, apis.extensible, apis.successors, apis.resource_hash
FROM api_definitions AS apis
         INNER JOIN (
--  select GAs
    SELECT a1.id, a1.tenant_id::text FROM applications AS a1
    UNION ALL
--  select CRM
    SELECT a.id, t.parent::text as tenant_id FROM applications AS a
                                                INNER JOIN  business_tenant_mappings AS t ON t.id = a.tenant_id WHERE t.parent IS NOT NULL
    UNION ALL
--  select SA
    SELECT l.app_id as id, asa.selector_value::text as tenant_id FROM labels AS l
                                                                          INNER JOIN automatic_scenario_assignments AS asa ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario AND asa.selector_key='global_subaccount_id'
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON apis.app_id = t_apps.id
         LEFT JOIN specifications as specs on apis.id = specs.api_def_id;

CREATE OR REPLACE VIEW tenants_events  AS
SELECT DISTINCT t_apps.tenant_id, events.id, events.app_id, events.name, events.description, events.group_name, events.version_value,
                events.version_deprecated, events.version_deprecated_since, events.version_for_removal, events.ord_id,
                events.short_description, events.system_instance_aware, events.changelog_entries, events.links, events.tags,
                events.countries, events.release_status, events.sunset_date, events.labels, events.package_id, events.visibility,
                events.disabled, events.part_of_products, events.line_of_business, events.industry, events.ready, events.created_at,
                events.updated_at, events.deleted_at, events.error, events.extensible, events.successors, events.resource_hash
FROM event_api_definitions AS events
         INNER JOIN (
--  select GAs
    SELECT a1.id, a1.tenant_id::text FROM applications AS a1
    UNION ALL
--  select CRM
    SELECT a.id, t.parent::text as tenant_id FROM applications AS a
                                                INNER JOIN  business_tenant_mappings AS t ON t.id = a.tenant_id WHERE t.parent IS NOT NULL
    UNION ALL
--  select SA
    SELECT l.app_id as id, asa.selector_value::text as tenant_id FROM labels AS l
                                                                          INNER JOIN automatic_scenario_assignments AS asa ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario AND asa.selector_key='global_subaccount_id'
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON events.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_bundles  AS
SELECT DISTINCT t_apps.tenant_id,b.id, b.app_id, b.name, b.description, b.instance_auth_request_json_schema,b. default_instance_auth,
                b.ord_id, b.short_description, b.links, b.labels, b.credential_exchange_strategies, b.ready, b.created_at,
                b.updated_at, b.deleted_at, b.error
FROM bundles AS b
         INNER JOIN (
--  select GAs
    SELECT a1.id, a1.tenant_id::text FROM applications AS a1
    UNION ALL
--  select CRM
    SELECT a.id, t.parent::text as tenant_id FROM applications AS a
                                                INNER JOIN  business_tenant_mappings AS t ON t.id = a.tenant_id WHERE t.parent IS NOT NULL
    UNION ALL
--  select SA
    SELECT l.app_id AS id, asa.selector_value::text AS tenant_id FROM labels AS l
                                                                          INNER JOIN automatic_scenario_assignments AS asa ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario AND asa.selector_key='global_subaccount_id'
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON b.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_specifications  AS
SELECT DISTINCT t_api_event_def.tenant_id, spec.id, spec.api_def_id, spec.event_def_id, spec.spec_data, spec.api_spec_format, spec.api_spec_type, spec.event_spec_format, spec.event_spec_type, spec.custom_type, spec.created_at
FROM specifications AS spec
         INNER JOIN (
--  select APIs
    SELECT a.id, a.tenant_id FROM tenants_apis AS a
    UNION ALL
--  select APIs
    SELECT e.id, e.tenant_id FROM tenants_events AS e
) AS t_api_event_def ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;

COMMIT;
