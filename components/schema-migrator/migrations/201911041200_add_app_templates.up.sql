
CREATE TYPE app_templates_access_level AS ENUM (
    'GLOBAL'
);

CREATE TABLE app_templates (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name varchar(256) NOT NULL,
    description text,
    application_input JSONB NOT NULL,
    placeholders JSONB,
    access_level app_templates_access_level NOT NULL
);

ALTER TABLE applications ADD COLUMN integration_system_id uuid;
ALTER TABLE applications
    ADD CONSTRAINT applications_integration_system_id_fk
        FOREIGN KEY (integration_system_id) REFERENCES integration_systems (id) ON DELETE RESTRICT;
