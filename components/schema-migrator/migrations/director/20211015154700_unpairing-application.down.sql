BEGIN;

DROP VIEW IF EXISTS tenants_apps;

ALTER TABLE applications
    ALTER COLUMN status_condition DROP DEFAULT;

ALTER TYPE application_status_condition RENAME TO application_status_condition_old;
CREATE TYPE application_status_condition AS ENUM (
    'INITIAL',
    'CONNECTED',
    'FAILED',
    'CREATING',
    'CREATE_FAILED',
    'CREATE_SUCCEEDED',
    'UPDATING',
    'UPDATE_FAILED',
    'UPDATE_SUCCEEDED',
    'DELETING',
    'DELETE_FAILED',
    'DELETE_SUCCEEDED',
    'MANAGED'
);

ALTER TABLE applications
    ALTER COLUMN status_condition TYPE application_status_condition
        USING status_condition::text::application_status_condition;

ALTER TABLE applications
    ALTER COLUMN status_condition
        SET DEFAULT 'INITIAL' ::application_status_condition;

DROP TYPE application_status_condition_old;

ALTER TABLE webhooks
    ALTER COLUMN type TYPE VARCHAR(255);

DROP TYPE webhook_type;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'UNREGISTER_APPLICATION',
    'OPEN_RESOURCE_DISCOVERY'
    );

ALTER TABLE webhooks
    ALTER COLUMN type TYPE webhook_type USING (type::webhook_type);

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, provider_tenant_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, product_type)
AS
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
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM applications a1
               UNION ALL
               SELECT a.id,
                      t.parent::text AS tenant_id,
                      t.parent::text AS provider_tenant_id
               FROM applications a
                        JOIN business_tenant_mappings t ON t.id = a.tenant_id
               WHERE t.parent IS NOT NULL
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id,(SELECT id::text FROM business_tenant_mappings WHERE external_tenant = a_s.tenant_id::text), cpr.provider_tenant::text AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? a_s.tenant_id::text) t_apps
              ON apps.id = t_apps.id;

COMMIT;
