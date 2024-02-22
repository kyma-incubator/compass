BEGIN;

-- Create table

CREATE TABLE IF NOT EXISTS tenant_business_types (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    code varchar(10) NOT NULL,
    name varchar(100) NOT NULL
);

-- Alter table for foreign keys

ALTER TABLE tenant_business_types
    ADD CONSTRAINT tenant_business_types_code_name_unique UNIQUE (code, name);

ALTER TABLE applications ADD COLUMN tenant_business_type_id uuid;

ALTER TABLE applications
    ADD CONSTRAINT applications_tenant_business_type_id_fk
        FOREIGN KEY (tenant_business_type_id) REFERENCES tenant_business_types (id) ON DELETE SET NULL;

-- Insert data into tenant_business_types table and setup foreign key relation with applications table

INSERT INTO tenant_business_types (id, name, code)
SELECT
    uuid_generate_v4(),
    l1.value AS name,
    l2.value AS code
FROM
    labels l1
        JOIN
    labels l2 ON l1.app_id = l2.app_id AND l1.key = 'tenantBusinessTypeName' AND l2.key = 'tenantBusinessTypeCode' GROUP BY name, code;

UPDATE applications a
SET tenant_business_type_id = subquery.tbt_id
FROM (
    SELECT a.id AS app_id, tbt.id AS tbt_id
    FROM applications a
    INNER JOIN labels l1 ON a.id = l1.app_id AND l1.key = 'tenantBusinessTypeCode'
    INNER JOIN labels l2 ON a.id = l2.app_id AND l2.key = 'tenantBusinessTypeName'
    INNER JOIN tenant_business_types tbt ON tbt.code = l1.value::TEXT AND tbt.name = l2.value::TEXT
) subquery
WHERE a.id = subquery.app_id;

-- Recreate views

DROP VIEW IF EXISTS listening_applications;
DROP VIEW IF EXISTS tenants_apps;

CREATE OR REPLACE VIEW listening_applications
            (id, app_template_id, system_number, local_tenant_id, name, description, status_condition, status_timestamp,
             system_status, healthcheck_url, integration_system_id, provider_name, base_url, application_namespace,
             labels, tags, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels, tenant_business_type_id,
             webhook_type)
AS
SELECT a.id,
       a.app_template_id,
       a.system_number,
       a.local_tenant_id,
       a.name,
       a.description,
       a.status_condition,
       a.status_timestamp,
       a.system_status,
       a.healthcheck_url,
       a.integration_system_id,
       a.provider_name,
       a.base_url,
       a.application_namespace,
       a.labels,
       a.tags,
       a.ready,
       a.created_at,
       a.updated_at,
       a.deleted_at,
       a.error,
       a.correlation_ids,
       a.documentation_labels,
       a.tenant_business_type_id,
       w.type
FROM applications a
         JOIN webhooks w on w.app_id = a.id or w.app_template_id = a.app_template_id;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, formation_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, tags, ready, created_at, updated_at, deleted_at,
             error, app_template_id, correlation_ids, system_number, application_namespace, local_tenant_id,
             tenant_business_type_id, product_type)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
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
                apps.tags,
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                COALESCE(apps.application_namespace, tmpl.application_namespace) AS application_namespace,
                apps.local_tenant_id,
                apps.tenant_business_type_id,
                tmpl.name                                                        AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON apps.id = t_apps.id;

COMMIT;