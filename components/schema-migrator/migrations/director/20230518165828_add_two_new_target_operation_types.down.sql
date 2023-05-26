BEGIN;

ALTER TABLE formation_constraints DROP CONSTRAINT formation_constraints_target_operation_check;
ALTER TABLE formation_constraints ADD CONSTRAINT formation_constraints_target_operation_check CHECK ( formation_constraints.target_operation in ('ASSIGN_FORMATION','UNASSIGN_FORMATION','CREATE_FORMATION','DELETE_FORMATION','GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION','GENERATE_FORMATION_NOTIFICATION','LOAD_FORMATIONS','SELECT_SYSTEMS_FOR_FORMATION') );

COMMIT;