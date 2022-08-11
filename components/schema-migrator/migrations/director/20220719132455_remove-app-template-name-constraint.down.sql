BEGIN;

ALTER TABLE app_templates
    ADD CONSTRAINT application_template_name_unique UNIQUE (name);

COMMIT;
