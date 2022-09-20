BEGIN;

ALTER TABLE app_templates
    DROP CONSTRAINT application_template_name_unique;

COMMIT;
