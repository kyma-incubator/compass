BEGIN;

INSERT INTO assignment_operations (id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp)
SELECT uuid_generate_v4(), 'UNASSIGN', fa.id, fa.formation_id, 'UNASSIGN_OBJECT', CURRENT_TIMESTAMP
FROM formation_assignments fa
WHERE fa.state IN ('INSTANCE_CREATOR_DELETING', 'INSTANCE_CREATOR_DELETE_ERROR');

COMMIT;
