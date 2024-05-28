BEGIN;

ALTER TABLE formation_assignments DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
ADD CONSTRAINT formation_assignments_state_check
CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETING', 'DELETE_ERROR', 'CONFIG_PENDING', 'INSTANCE_CREATOR_DELETING', 'INSTANCE_CREATOR_DELETE_ERROR'));

---

UPDATE formation_assignments
SET state = 'INSTANCE_CREATOR_DELETING'
WHERE id IN (SELECT formation_assignment_id FROM assignment_operations where type = 'INSTANCE_CREATOR_UNASSING') AND state = 'DELETING';

UPDATE formation_assignments
SET state = 'INSTANCE_CREATOR_DELETE_ERROR'
WHERE id IN (SELECT formation_assignment_id FROM assignment_operations where type = 'INSTANCE_CREATOR_UNASSING') AND state = 'DELETE_ERROR';

---

DELETE FROM assignment_operations
WHERE type = 'INSTANCE_CREATOR_UNASSIGN';

ALTER TABLE assignment_operations DROP CONSTRAINT assignment_operations_type_check;

ALTER TABLE assignment_operations
ADD CONSTRAINT assignment_operations_type_check
CHECK (type IN ('ASSIGN', 'UNASSIGN'));

COMMIT;
