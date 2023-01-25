BEGIN;

-- Drop views that use the webhook type as dependency

DROP VIEW IF EXISTS application_webhooks_tenants;
DROP VIEW IF EXISTS runtime_webhooks_tenants;
DROP VIEW IF EXISTS webhooks_tenants;

-- Drop constraint so that it can be updated

ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhook_owner_id_unique;

-- Add formation_template_id to webhooks with respective constraints and index

ALTER TABLE webhooks ADD COLUMN formation_template_id UUID;

ALTER TABLE webhooks ADD CONSTRAINT webhooks_formation_template_id_fk FOREIGN KEY (formation_template_id) REFERENCES formation_templates(id) ON DELETE CASCADE;

ALTER TABLE webhooks ADD CONSTRAINT webhook_formation_template_id_type_unique UNIQUE (formation_template_id, type);

CREATE INDEX webhooks_formation_template_id ON webhooks(formation_template_id) WHERE webhooks.formation_template_id IS NOT NULL;

ALTER TABLE webhooks ADD CONSTRAINT webhook_owner_id_unique
    CHECK ((app_template_id IS NOT NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS NULL AND formation_template_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NOT NULL AND runtime_id IS NULL AND integration_system_id IS NULL AND formation_template_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NULL AND runtime_id IS NOT NULL AND integration_system_id IS NULL AND formation_template_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS NOT NULL AND formation_template_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS NULL AND formation_template_id IS NOT NULL));


-- Add FORMATION_LIFECYCLE webhook_type

ALTER TABLE webhooks ALTER COLUMN type TYPE VARCHAR(255);

DROP TYPE webhook_type;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'UNREGISTER_APPLICATION',
    'OPEN_RESOURCE_DISCOVERY',
    'APPLICATION_TENANT_MAPPING',
    'FORMATION_LIFECYCLE'
    );


ALTER TABLE webhooks
    ALTER COLUMN type TYPE webhook_type USING (type::webhook_type);

-- Create new view for tenant access checks used for entity with embedded tenant (formation templates) and its child (webhook)

CREATE OR REPLACE VIEW formation_templates_webhooks_tenants (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
                                                             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
                                                             app_template_id, formation_template_id, tenant_id, owner)
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
       w.formation_template_id,
       ft.tenant_id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id;


-- Recreate views

CREATE OR REPLACE VIEW webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, formation_template_id, tenant_id, owner)
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
       w.formation_template_id,
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
       w.formation_template_id,
       tr.tenant_id,
       tr.owner
FROM webhooks w
         JOIN tenant_runtimes tr ON w.runtime_id = tr.id
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
       w.formation_template_id,
       ft.tenant_id,
       true
FROM webhooks w
        JOIN formation_templates ft on w.formation_template_id = ft.id;


CREATE OR REPLACE VIEW runtime_webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, formation_template_id, tenant_id, owner)
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
       w.formation_template_id,
       tr.tenant_id,
       tr.owner
FROM webhooks w
         JOIN tenant_runtimes tr ON w.runtime_id = tr.id;

CREATE OR REPLACE VIEW application_webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, formation_template_id, tenant_id, owner)
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
       w.formation_template_id,
       ta.tenant_id,
       ta.owner
FROM webhooks w
         JOIN tenant_applications ta ON w.app_id = ta.id;

COMMIT;
