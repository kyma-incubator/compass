BEGIN;

ALTER TABLE applications DROP CONSTRAINT applications_app_template_id_fkey;
ALTER TABLE applications
    ADD CONSTRAINT applications_app_template_id_fkey
        FOREIGN KEY (app_template_id) REFERENCES app_templates(id) ON DELETE RESTRICT;

COMMIT;