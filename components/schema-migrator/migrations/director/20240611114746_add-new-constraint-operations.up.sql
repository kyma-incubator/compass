BEGIN;

ALTER TABLE formation_constraints DROP CONSTRAINT formation_constraints_target_operation_check;

COMMIT;
