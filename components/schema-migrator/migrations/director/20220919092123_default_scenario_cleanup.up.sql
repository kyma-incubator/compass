BEGIN;

-- delete all DEFAULT formations - improperly migrated
DELETE FROM formations WHERE name = 'DEFAULT';

COMMIT;
