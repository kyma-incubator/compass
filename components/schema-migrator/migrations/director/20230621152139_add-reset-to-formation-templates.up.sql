BEGIN;

ALTER TABLE formation_templates
ADD COLUMN supports_reset bool not null default false;

COMMIT;
