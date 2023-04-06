BEGIN;

ALTER TABLE formation_constraints DROP CONSTRAINT formation_constraints_constraint_type_check;
ALTER TABLE formation_constraints ADD CONSTRAINT formation_constraints_constraint_type_check CHECK ( formation_constraints.constraint_type IN ('PRE', 'POST'));

ALTER TABLE formation_constraints DROP CONSTRAINT formation_constraints_target_operation_check;
ALTER TABLE formation_constraints ADD CONSTRAINT formation_constraints_target_operation_check CHECK ( target_operation in ('ASSIGN_FORMATION','UNASSIGN_FORMATION','CREATE_FORMATION','DELETE_FORMATION','GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION','GENERATE_FORMATION_NOTIFICATION') );

COMMIT;
