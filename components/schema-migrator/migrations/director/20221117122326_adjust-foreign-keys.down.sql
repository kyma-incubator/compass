BEGIN;

ALTER TABLE formations DROP CONSTRAINT formations_formation_template_id_fkey;
ALTER TABLE formations ADD CONSTRAINT formations_formation_template_id_fkey
    FOREIGN KEY (formation_template_id) REFERENCES formation_templates(id) ON DELETE CASCADE;

ALTER TABLE formation_assignments DROP CONSTRAINT formation_assignments_formation_id_only_fk;
ALTER TABLE formation_assignments DROP CONSTRAINT formation_assignments_tenant_id_fk;

COMMIT;
