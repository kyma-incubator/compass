BEGIN;

ALTER TABLE formation_template_constraint_references DROP CONSTRAINT IF EXISTS unique_template_id_constraint_id;

ALTER TABLE formation_constraints ALTER COLUMN description TYPE varchar(255);

COMMIT;