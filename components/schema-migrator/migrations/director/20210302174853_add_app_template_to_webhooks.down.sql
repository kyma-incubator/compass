BEGIN;

ALTER TABLE webhooks
    DROP CONSTRAINT IF EXISTS webhook_app_id_type_unique,
    DROP CONSTRAINT IF EXISTS webhook_app_template_id_type_unique,
    DROP CONSTRAINT IF EXISTS webhook_runtime_id_type_unique,
    DROP CONSTRAINT IF EXISTS webhook_integration_system_id_type_unique,
    DROP CONSTRAINT IF EXISTS webhook_owner_id_unique,
    ALTER COLUMN tenant_id SET NOT NULL,
    DROP CONSTRAINT IF EXISTS webhooks_app_template_id_fkey,
    DROP COLUMN IF EXISTS app_template_id;

UPDATE applications
SET app_template_id = NULL
WHERE id = id;

ALTER TABLE applications
    DROP CONSTRAINT IF EXISTS applications_app_template_id_fkey,
    DROP COLUMN app_template_id;

COMMIT;