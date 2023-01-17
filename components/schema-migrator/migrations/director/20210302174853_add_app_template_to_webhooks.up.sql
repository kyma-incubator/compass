BEGIN;

ALTER TABLE webhooks
    ADD COLUMN app_template_id uuid,
    ADD CONSTRAINT webhooks_app_template_id_fkey FOREIGN KEY (app_template_id) REFERENCES app_templates(id) ON DELETE CASCADE,
    ALTER COLUMN tenant_id DROP NOT NULL,
    ADD CONSTRAINT webhook_owner_id_unique
        CHECK ((app_template_id IS NOT NULL AND tenant_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS NULL)
            OR (app_template_id IS NULL AND tenant_id IS NOT NULL AND app_id IS NOT NULL AND runtime_id IS NULL AND integration_system_id IS NULL)
            OR (app_template_id IS NULL AND tenant_id IS NOT NULL AND app_id IS NULL AND runtime_id IS NOT NULL AND integration_system_id IS NULL)
            OR (app_template_id IS NULL AND tenant_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS  NOT NULL)),
    ADD CONSTRAINT webhook_app_id_type_unique UNIQUE (app_id, type),
    ADD CONSTRAINT webhook_app_template_id_type_unique UNIQUE (app_template_id, type),
    ADD CONSTRAINT webhook_runtime_id_type_unique UNIQUE (runtime_id, type),
    ADD CONSTRAINT webhook_integration_system_id_type_unique UNIQUE (integration_system_id, type);


ALTER TABLE applications
    ADD COLUMN app_template_id uuid;

ALTER TABLE applications
    ADD CONSTRAINT applications_app_template_id_fkey FOREIGN KEY (app_template_id) REFERENCES app_templates(id) ON DELETE SET NULL;

UPDATE applications
SET app_template_id = (SELECT id FROM app_templates
                       WHERE to_jsonb(name) = (SElECT value FROM labels
                                               WHERE app_id = applications.id AND key = 'applicationType'))
WHERE id = id;

COMMIT;
