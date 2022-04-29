BEGIN;

ALTER TABLE labels
    ADD COLUMN app_template_id UUID;

ALTER TABLE labels
    ADD CONSTRAINT labels_app_template_id_fk FOREIGN KEY (app_template_id) REFERENCES app_templates(id) ON DELETE CASCADE;

COMMIT;
