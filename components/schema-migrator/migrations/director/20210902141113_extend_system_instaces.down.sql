BEGIN;

DROP VIEW tenants_apps;

CREATE OR REPLACE VIEW tenants_apps  AS
SELECT DISTINCT t_apps.tenant_id, apps.id, apps.name, apps.description, apps.status_condition,
                apps.status_timestamp, apps.healthcheck_url, apps.integration_system_id,
                apps.provider_name, apps.base_url, apps.labels, apps.ready, apps.created_at,
                apps.updated_at, apps.deleted_at, apps.error, apps.app_template_id, apps.correlation_ids
FROM applications AS apps
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
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON apps.id = t_apps.id;

COMMIT;
