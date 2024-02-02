BEGIN;

ALTER TABLE formation_template_constraint_references ADD CONSTRAINT unique_template_id_constraint_id UNIQUE (formation_template_id, formation_constraint_id);

ALTER TABLE formation_constraints ALTER COLUMN description TYPE varchar(2048);

COMMIT;
