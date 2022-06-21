BEGIN;

ALTER TABLE formation_templates ADD CONSTRAINT formation_template_unique_name UNIQUE (name);

COMMIT;
