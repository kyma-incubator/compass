BEGIN;

ALTER TABLE formation_assignments
    ADD COLUMN last_operation_initiator UUID,
    ADD COLUMN last_operation_initiator_type VARCHAR(256),
    ADD COLUMN last_operation VARCHAR(256);

ALTER TABLE formation_assignments
    ADD CONSTRAINT last_operation_initiator_type_check CHECK (last_operation_initiator_type IN ('APPLICATION', 'RUNTIME', 'RUNTIME_CONTEXT')),
    ADD CONSTRAINT last_operation_check CHECK (last_operation IN ('assign', 'unassign'));

ALTER TABLE formation_assignments
    DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
    ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETING', 'DELETE_ERROR', 'CONFIG_PENDING'));

UPDATE formation_assignments SET last_operation = 'assign', last_operation_initiator = source, last_operation_initiator_type = source_type;

COMMIT;
