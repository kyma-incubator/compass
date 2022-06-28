BEGIN;

DELETE FROM labels WHERE key = 'runtimeType';

COMMIT;
