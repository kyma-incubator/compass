BEGIN;

CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$

DECLARE
data json;
    notification json;

BEGIN

    -- Convert the old or new row to JSON, based on the kind of action.
    -- Action = DELETE?             -> OLD row
    -- Action = INSERT or UPDATE?   -> NEW row
    IF (TG_OP = 'DELETE') THEN
        data = row_to_json(OLD);
ELSE
        data = row_to_json(NEW);
END IF;

    -- Contruct the notification as a JSON string.
    notification = json_build_object(
            'table',TG_TABLE_NAME,
            'action', TG_OP,
            'data', data);


    -- Execute pg_notify(channel, notification)
    PERFORM pg_notify('events',notification::text);

    -- Result is ignored since this is an AFTER trigger
RETURN NULL;
END;

$$ LANGUAGE plpgsql;

CREATE TRIGGER formations_notify_event
    AFTER INSERT OR UPDATE OR DELETE ON formations
    FOR EACH ROW EXECUTE PROCEDURE notify_event();

CREATE TRIGGER formation_assignments_notify_event
    AFTER INSERT OR UPDATE OR DELETE ON formation_assignments
    FOR EACH ROW EXECUTE PROCEDURE notify_event();

-- TODO: this is needed only locally; do no execute the sql below on Canary as it will be done by a migration previously

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
    'SYNC',
    'ASYNC_CALLBACK'
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