BEGIN;

ALTER TABLE formation_templates ADD COLUMN tenant_id uuid;

ALTER TABLE formation_templates
    ADD CONSTRAINT formation_templates_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS formation_templates_tenant_id_idx ON formation_templates(tenant_id);

ALTER TABLE formation_templates DROP CONSTRAINT formation_template_unique_name;
ALTER TABLE formation_templates ADD CONSTRAINT formation_template_unique_name_and_tenant_id UNIQUE (name, tenant_id);

COMMIT;
