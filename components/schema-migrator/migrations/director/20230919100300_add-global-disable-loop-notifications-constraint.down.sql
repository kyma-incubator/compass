BEGIN;

DELETE FROM formation_constraints fc
WHERE fc.name = 'DoNotGenerateFormationAssignmentNotificationForLoopsGlobalApplication';

COMMIT;
