BEGIN;

ALTER TABLE formation_assignments
    DROP CONSTRAINT last_operation_initiator_type_check,
    DROP CONSTRAINT last_operation_check;

ALTER TABLE formation_assignments
    DROP COLUMN last_operation_initiator,
    DROP COLUMN last_operation_initiator_type,
    DROP COLUMN last_operation;

ALTER TABLE formation_assignments
    DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
    ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETE_ERROR', 'CONFIG_PENDING'));

COMMIT;