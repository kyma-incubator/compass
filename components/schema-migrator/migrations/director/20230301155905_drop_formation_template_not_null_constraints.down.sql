BEGIN;

ALTER TABLE formation_templates ALTER COLUMN runtime_type_display_name SET NOT NULL;
ALTER TABLE formation_templates ALTER COLUMN runtime_artifact_kind SET NOT NULL;
ALTER TABLE formation_templates ALTER COLUMN runtime_types SET NOT NULL;

COMMIT;
