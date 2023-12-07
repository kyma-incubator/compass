BEGIN;


ALTER TABLE formation_assignments
    DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
    ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETING', 'DELETE_ERROR', 'CONFIG_PENDING', 'INSTANCE_CREATOR_DELETING', 'INSTANCE_CREATOR_DELETE_ERROR'));

COMMIT;
