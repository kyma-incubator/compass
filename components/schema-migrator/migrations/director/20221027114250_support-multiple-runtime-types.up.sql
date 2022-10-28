BEGIN;

ALTER TABLE formation_templates ALTER COLUMN runtime_type TYPE jsonb USING to_jsonb(array[runtime_type])
ALTER TABLE formation_templates RENAME COLUMN runtime_type TO runtime_types;

COMMIT;
