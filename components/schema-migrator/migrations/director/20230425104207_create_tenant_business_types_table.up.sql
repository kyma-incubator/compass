BEGIN;

-- Drop views --
DROP VIEW IF EXISTS ord_tags_applications;
DROP VIEW IF EXISTS listening_applications;
DROP VIEW IF EXISTS tenants_apps;

-- Create tables --
CREATE TABLE tenant_business_types (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    code varchar(10) NOT NULL,
    name varchar(100) NOT NULL
);

-- Alter tables --
ALTER TABLE tenant_business_types
    ADD CONSTRAINT tenant_business_types_code_name_unique UNIQUE (code, name);

ALTER TABLE applications ADD COLUMN tenant_business_type_id uuid;

ALTER TABLE applications
    ADD CONSTRAINT applications_tenant_business_type_id_fk
        FOREIGN KEY (tenant_business_type_id) REFERENCES tenant_business_types (id) ON DELETE SET NULL;


-- Recreate views --
CREATE VIEW ord_tags_applications AS
SELECT id                  AS application_id,
       elements.value      AS value
FROM applications,
     jsonb_array_elements_text(applications.tags) AS elements;

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
             integration_system_id, provider_name, base_url, labels, tags, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, application_namespace, local_tenant_id, tenant_business_type_id, product_type)
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
                coalesce(apps.application_namespace, tmpl.application_namespace),
                apps.local_tenant_id,
                apps.tenant_business_type_id,
                tmpl.name AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      apps_subaccounts.formation_id
               FROM apps_subaccounts
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps
              ON apps.id = t_apps.id;

COMMIT;