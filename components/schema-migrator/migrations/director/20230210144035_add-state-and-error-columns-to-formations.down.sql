BEGIN;

ALTER TABLE formations
    DROP COLUMN state,
    DROP COLUMN error;

ALTER TABLE formation_constraints
    DROP CONSTRAINT formation_constraints_target_operation_check;

ALTER TABLE formation_constraints
    ADD CONSTRAINT formation_constraints_target_operation_check CHECK ( target_operation in ('ASSIGN_FORMATION','UNASSIGN_FORMATION','CREATE_FORMATION','DELETE_FORMATION','GENERATE_NOTIFICATION') );

COMMIT;
