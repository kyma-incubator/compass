BEGIN;

-- the only way to remove a cascade is to delete the constraint and recreate it
ALTER TABLE formation_template_constraint_references DROP CONSTRAINT formation_template_constraint_refe_formation_constraint_id_fkey;
ALTER TABLE formation_template_constraint_references ADD CONSTRAINT formation_template_constraint_refe_formation_constraint_id_fkey FOREIGN KEY (formation_constraint_id) REFERENCES formation_constraints(id);

COMMIT;
