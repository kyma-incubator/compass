BEGIN;
ALTER TABLE formation_constraints DROP CONSTRAINT formation_constraints_constraint_type_check;
ALTER TABLE formation_constraints ADD CONSTRAINT formation_constraints_constraint_type_check CHECK ( formation_constraints.constraint_type IN ('PRE', 'POST'));
COMMIT;
