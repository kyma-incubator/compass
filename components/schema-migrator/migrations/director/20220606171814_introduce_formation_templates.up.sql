BEGIN;

CREATE TABLE formation_templates
(
    id                               uuid NOT NULL,
    name                             varchar(256) NOT NULL,
    application_types                jsonb NOT NULL,
    runtime_types                    jsonb NOT NULL,
    missing_artifact_info_message    text NOT NULL,
    missing_artifact_warning_message text NOT NULL,

    PRIMARY KEY(id)
);


INSERT INTO formation_templates(id, name, application_types, runtime_types, missing_artifact_info_message, missing_artifact_warning_message)
VALUES (uuid_generate_v4(), 'Side-by-side extensibility with Kyma',
        '["SAP Cloud for Customer", "SAP Commerce Cloud", "SAP Field Service Management", "SAP Marketing Cloud"]',
        '["kyma"]',
        'Subaccount %s is missing a required SAP BTP Kyma environment instance. You can proceed but it must be created later.',
        'Subaccount %s is missing a required SAP BTP Kyma environment instance. You must create it.'
        );

COMMIT;
