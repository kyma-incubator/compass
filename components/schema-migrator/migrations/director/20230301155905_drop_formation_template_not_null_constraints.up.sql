BEGIN;

ALTER TABLE formation_templates ALTER COLUMN runtime_type_display_name DROP NOT NULL;
ALTER TABLE formation_templates ALTER COLUMN runtime_artifact_kind DROP NOT NULL;

COMMIT;
