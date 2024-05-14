BEGIN;

DELETE FROM assignment_operations
WHERE formation_assignment_id IN (SELECT id FROM formation_assignments WHERE state IN ('INSTANCE_CREATOR_DELETING', 'INSTANCE_CREATOR_DELETE_ERROR'));

COMMIT;
