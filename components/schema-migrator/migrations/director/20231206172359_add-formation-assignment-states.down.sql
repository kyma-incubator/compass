BEGIN;

UPDATE formation_assignments
    SET state = 'DELETING'
    WHERE state = 'INSTANCE_CREATOR_DELETING';
UPDATE formation_assignments
    SET state = 'DELETE_ERROR'
    WHERE state = 'INSTANCE_CREATOR_DELETE_ERROR';

ALTER TABLE formation_assignments
    DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
    ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETING', 'DELETE_ERROR', 'CONFIG_PENDING'));

COMMIT;
