BEGIN;

-- Drop views --
DROP VIEW IF EXISTS listening_applications;

-- Recreate views --
CREATE OR REPLACE VIEW listening_applications
            (id, app_template_id, system_number, local_tenant_id, name, description, status_condition, status_timestamp,
             system_status, healthcheck_url, integration_system_id, provider_name, base_url, application_namespace,
             labels, tags, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels,
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
       w.type
FROM applications a
         JOIN webhooks w on w.app_id = a.id or w.app_template_id = a.app_template_id;

-- Alter tables --
ALTER TABLE applications DROP COLUMN IF EXISTS tenant_business_type_id;

ALTER TABLE applications
    DROP CONSTRAINT IF EXISTS applications_tenant_business_type_id_fk;

-- Drop tables --
DROP TABLE IF EXISTS tenant_business_types;

COMMIT;