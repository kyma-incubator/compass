BEGIN;

ALTER TABLE webhooks
    DROP CONSTRAINT IF EXISTS app_or_template,
    ALTER COLUMN tenant_id SET NOT NULL,
    ALTER COLUMN app_id SET NOT NULL,
    DROP CONSTRAINT IF EXISTS webhooks_app_template_id_fkey,
    DROP COLUMN IF EXISTS app_template_id;


ALTER TABLE applications
    DROP CONSTRAINT IF EXISTS applications_app_template_id_fkey,
    DROP COLUMN app_template_id;

COMMIT;