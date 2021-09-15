BEGIN;

CREATE OR REPLACE FUNCTION apps_subaccounts_func()
    RETURNS TABLE
            (
                id                 uuid,
                tenant_id          uuid,
                provider_tenant_id uuid
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT l.app_id                 AS id,
               asa.selector_value::uuid AS tenant_id,
               asa.selector_value::uuid AS provider_tenant_id
        FROM labels l
                 -- 2) Get subaccounts in those scenarios (Putting a subaccount in a
                 -- scenario will reflect on creating an ASA for the subaccount.
                 JOIN automatic_scenario_assignments asa
                      ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario::text AND
                         asa.selector_key::text = 'global_subaccount_id'::text
             -- 1) Get all scenario labels for applications
        WHERE l.app_id IS NOT NULL
          AND l.key::text = 'scenarios'::text;
END
$$;

CREATE OR REPLACE FUNCTION consumers_provider_for_runtimes_func()
    RETURNS TABLE
            (
                provider_tenant  jsonb,
                consumer_tenants jsonb
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT l1.value AS provider_tenant, l2.value AS consumer_tenants
        FROM (SELECT * FROM labels WHERE key::text = 'global_subaccount_id') l1 -- Get the subaccount for each runtime
                 JOIN (SELECT * FROM labels WHERE key::text = 'consumer_subaccount_ids') l2 -- Get all the consumer subaccounts for each runtime
                      ON l1.runtime_id = l2.runtime_id AND l1.runtime_id IS NOT NULL;
END
$$;

DROP VIEW IF EXISTS tenants_apps;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, provider_tenant_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, product_type)
AS
WITH apps_subaccounts AS (SELECT * FROM apps_subaccounts_func()),
     consumers_provider_for_runtimes AS (SELECT * FROM consumers_provider_for_runtimes_func())
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                      a1.tenant_id,
                      a1.tenant_id AS provider_tenant_id
               FROM applications a1
               UNION ALL
               SELECT a.id,
                      t.parent AS tenant_id,
                      t.parent AS provider_tenant_id
               FROM applications a
                        JOIN business_tenant_mappings t ON t.id = a.tenant_id
               WHERE t.parent IS NOT NULL
               UNION ALL
               SELECT *
               FROM apps_subaccounts
               UNION ALL
               SELECT a_s.id, a_s.tenant_id, (cpr.provider_tenant #>> '{}')::uuid AS provider_tenant_id
               FROM apps_subaccounts a_s
                        JOIN consumers_provider_for_runtimes cpr
                             ON cpr.consumer_tenants ? a_s.tenant_id::text) t_apps
              ON apps.id = t_apps.id;

COMMIT;