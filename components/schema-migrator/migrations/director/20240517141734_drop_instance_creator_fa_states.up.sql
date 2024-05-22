BEGIN;

ALTER TABLE assignment_operations DROP CONSTRAINT assignment_operations_type_check;

ALTER TABLE assignment_operations
ADD CONSTRAINT assignment_operations_type_check
CHECK (type IN ('ASSIGN', 'UNASSIGN', 'INSTANCE_CREATOR_UNASSIGN'));

---

INSERT INTO assignment_operations (id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp)
SELECT uuid_generate_v4(), 'INSTANCE_CREATOR_UNASSIGN', fa.id, fa.formation_id, 'UNASSIGN_OBJECT', CURRENT_TIMESTAMP
FROM formation_assignments fa
WHERE fa.state IN ('INSTANCE_CREATOR_DELETING', 'INSTANCE_CREATOR_DELETE_ERROR');

---

UPDATE formation_assignments
SET state = 'DELETING'
WHERE state = 'INSTANCE_CREATOR_DELETING';

UPDATE formation_assignments
SET state = 'DELETE_ERROR'
WHERE state = 'INSTANCE_CREATOR_DELETE_ERROR';

---

ALTER TABLE formation_assignments DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETING', 'DELETE_ERROR', 'CONFIG_PENDING'));

COMMIT;
