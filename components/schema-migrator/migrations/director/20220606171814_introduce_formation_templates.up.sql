BEGIN;

CREATE TYPE artifact_kind AS ENUM (
    'SUBSCRIPTION',
    'SERVICE_INSTANCE',
    'ENVIRONMENT_INSTANCE'
);

CREATE TABLE formation_templates
(
    id                               uuid NOT NULL,
    name                             varchar(256) NOT NULL,
    application_types                jsonb NOT NULL,
    runtime_type                     varchar(256) NOT NULL,
    runtime_type_display_name        varchar(256) NOT NULL,
    runtime_artifact_kind            artifact_kind NOT NULL,
    PRIMARY KEY(id)
);


INSERT INTO formation_templates(id, name, application_types, runtime_type, runtime_type_display_name, runtime_artifact_kind)
VALUES (uuid_generate_v4(), 'Side-by-side extensibility with Kyma',
        '["SAP Cloud for Customer", "SAP Commerce Cloud", "SAP Field Service Management", "SAP Marketing Cloud"]',
        'kyma',
        'SAP BTP Kyma',
        'ENVIRONMENT_INSTANCE'
        );

COMMIT;
