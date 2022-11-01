BEGIN;

ALTER TABLE formation_templates ALTER COLUMN runtime_types TYPE varchar(256) USING runtime_types->>0;
ALTER TABLE formation_templates RENAME COLUMN runtime_types TO runtime_type;

COMMIT;
