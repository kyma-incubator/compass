BEGIN;

CREATE OR REPLACE VIEW tenants_apis  AS
SELECT DISTINCT t_apps.tenant_id, apis.id, apis.app_id, apis.name, apis.description, apis.group_name, apis.default_auth,
                apis.version_value, apis.version_deprecated, apis.version_deprecated_since, apis.version_for_removal,
                apis.ord_id, apis.short_description, apis.system_instance_aware, apis.api_protocol, apis.tags,
                apis.countries, apis.links, apis.api_resource_links, apis.release_status, apis.sunset_date,
                apis.changelog_entries, apis.labels, apis.package_id, apis.visibility, apis.disabled, apis.part_of_products,
                apis.line_of_business, apis.industry, apis.ready, apis.created_at, apis.updated_at, apis.deleted_at, apis.error,
                apis.implementation_standard, apis.custom_implementation_standard, apis.custom_implementation_standard_description,
                apis.target_urls, apis.extensible, apis.successors, apis.resource_hash
FROM api_definitions AS apis
         INNER JOIN (
--  select GAs
    SELECT a1.id, a1.tenant_id FROM applications AS a1
    UNION ALL
--  select CRM
    SELECT a.id, t.parent as tenant_id FROM applications AS a
                                                INNER JOIN  business_tenant_mappings AS t ON t.id = a.tenant_id WHERE t.parent IS NOT NULL
    UNION ALL
--  select SA
    SELECT l.app_id as id, asa.selector_value::uuid as tenant_id FROM labels AS l
                                                                          INNER JOIN automatic_scenario_assignments AS asa ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario AND asa.selector_key='global_subaccount_id'
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON apis.app_id = t_apps.id;

COMMIT;
