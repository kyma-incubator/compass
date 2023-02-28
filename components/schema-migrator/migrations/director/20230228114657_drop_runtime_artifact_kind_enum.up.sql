BEGIN;

ALTER TABLE formation_templates
    ALTER COLUMN runtime_artifact_kind
        SET DATA TYPE varchar(256) USING runtime_artifact_kind::text;

DROP TYPE artifact_kind;

COMMIT;
