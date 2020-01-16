DROP TABLE app_templates;

DROP TYPE app_templates_access_level;

ALTER TABLE applications
    DROP CONSTRAINT applications_integration_system_id_fk;
ALTER TABLE applications DROP COLUMN integration_system_id;
