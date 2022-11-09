BEGIN;

ALTER TABLE formation_assignments
    DROP CONSTRAINT last_operation_initiator_type_check,
    DROP CONSTRAINT last_operation_check;

ALTER TABLE formation_assignments
    DROP COLUMN last_operation_initiator,
    DROP COLUMN last_operation_initiator_type,
    DROP COLUMN last_operation;

ALTER TABLE formation_assignments
    DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
    ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETE_ERROR', 'CONFIG_PENDING'));


-- Drop views that use the webhook type as dependency
DROP VIEW IF EXISTS application_webhooks_tenants;
DROP VIEW IF EXISTS runtime_webhooks_tenants;
DROP VIEW IF EXISTS webhooks_tenants;

-- Re-create webhook type with the new ASYNC_CALLBACK value
ALTER TABLE webhooks ALTER COLUMN mode DROP DEFAULT;
ALTER TABLE webhooks ALTER COLUMN mode TYPE VARCHAR(256);

DROP TYPE webhook_mode;

CREATE TYPE webhook_mode AS ENUM (
    'ASYNC',
    'SYNC'
);

ALTER TABLE webhooks ALTER COLUMN mode TYPE webhook_mode USING (mode::webhook_mode);
ALTER TABLE webhooks ALTER COLUMN mode SET DEFAULT 'SYNC';


-- Re-create views
CREATE OR REPLACE VIEW application_webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, tenant_id, owner)
AS
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       ta.tenant_id,
       ta.owner
FROM webhooks w
         JOIN tenant_applications ta ON w.app_id = ta.id;


CREATE OR REPLACE VIEW runtime_webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, tenant_id, owner)
AS
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       tr.tenant_id,
       tr.owner
FROM webhooks w
         JOIN tenant_runtimes tr ON w.runtime_id = tr.id;


CREATE OR REPLACE VIEW webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, tenant_id, owner)
AS
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       ta.tenant_id,
       ta.owner
FROM webhooks w
         JOIN tenant_applications ta ON w.app_id = ta.id
UNION ALL
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       tr.tenant_id,
       tr.owner
FROM webhooks w
         JOIN tenant_runtimes tr ON w.runtime_id = tr.id;

COMMIT;
