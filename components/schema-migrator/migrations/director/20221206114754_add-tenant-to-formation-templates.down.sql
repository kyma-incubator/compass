BEGIN;

ALTER TABLE formation_templates DROP CONSTRAINT formation_template_unique_name_and_tenant_id;
ALTER TABLE formation_templates ADD CONSTRAINT formation_template_unique_name UNIQUE (name);

DROP INDEX IF EXISTS formation_templates_tenant_id_idx;

ALTER TABLE formation_templates DROP CONSTRAINT formation_templates_tenant_fk;

ALTER TABLE formation_templates DROP COLUMN tenant_id;

COMMIT;
