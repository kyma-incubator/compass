BEGIN;

ALTER TABLE webhooks
    ADD COLUMN app_template_id uuid,
    ADD FOREIGN KEY (app_template_id) REFERENCES app_templates(id) ON DELETE CASCADE,
    ALTER COLUMN app_id DROP NOT NULL,
    ALTER COLUMN tenant_id DROP NOT NULL,
    ADD CONSTRAINT app_or_template
        CHECK ( ((app_id IS NULL AND tenant_id IS NULL)  OR app_template_id IS NULL)
                    AND NOT((app_id IS NULL AND tenant_id IS NULL) AND app_template_id IS NULL));


ALTER TABLE applications
    ADD COLUMN app_template_id uuid,
    ADD FOREIGN KEY (app_template_id) REFERENCES app_templates(id) ON DELETE SET NULL;

--TODO: Fill app_template_id column for apps which were already created by templates

COMMIT;
