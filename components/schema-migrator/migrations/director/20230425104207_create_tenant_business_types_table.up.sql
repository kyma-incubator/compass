BEGIN;

-- Drop views --
DROP VIEW IF EXISTS listening_applications;

-- Recreate views --
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

COMMIT;