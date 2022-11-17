BEGIN;

-- Remove ON DELETE CASCADE
ALTER TABLE formations DROP CONSTRAINT formations_formation_template_id_fkey;
ALTER TABLE formations ADD CONSTRAINT formations_formation_template_id_fkey
    FOREIGN KEY (formation_template_id) REFERENCES formation_templates(id);

ALTER TABLE formation_assignments ADD CONSTRAINT formation_assignments_formation_id_only_fk
    FOREIGN KEY (formation_id) REFERENCES formations(id) ON DELETE CASCADE;

ALTER TABLE formation_assignments ADD CONSTRAINT formation_assignments_tenant_id_fk
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;

COMMIT;
