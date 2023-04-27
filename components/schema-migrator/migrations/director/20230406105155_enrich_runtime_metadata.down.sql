BEGIN;

ALTER TABLE runtimes
DROP COLUMN application_namespace;

DELETE FROM labels
WHERE runtime_id IN (SELECT runtime_id FROM labels WHERE key = 'global_subaccount_id') AND key = 'region';

COMMIT;
