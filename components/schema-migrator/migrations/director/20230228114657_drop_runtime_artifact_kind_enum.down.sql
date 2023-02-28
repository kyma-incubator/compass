BEGIN;

CREATE TYPE artifact_kind AS ENUM (
    'SUBSCRIPTION',
    'SERVICE_INSTANCE',
    'ENVIRONMENT_INSTANCE'
);

ALTER TABLE formation_templates
    ALTER COLUMN runtime_artifact_kind
        SET DATA TYPE artifact_kind USING runtime_artifact_kind::artifact_kind;

COMMIT;
