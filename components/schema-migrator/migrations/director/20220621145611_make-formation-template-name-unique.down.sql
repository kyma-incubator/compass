BEGIN;

ALTER TABLE formation_templates DROP CONSTRAINT formation_template_unique_name;

COMMIT;
