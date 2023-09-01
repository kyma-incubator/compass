BEGIN;

DELETE FROM formation_constraints fc
WHERE fc.operator = 'DoNotGenerateFormationAssignmentNotificationForLoops';

COMMIT;
