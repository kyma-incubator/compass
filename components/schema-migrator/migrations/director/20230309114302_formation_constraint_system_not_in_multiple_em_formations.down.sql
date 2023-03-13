BEGIN;

UPDATE formation_constraints
SET name = 'SubaccountInAtMostOneEventMeshFormation'
WHERE name = 'SubaccountInAtMostOneFormationOfGivenType';

DELETE FROM formation_constraints WHERE name = 'SystemInAtMostOneFormationOfGivenType';

COMMIT;
