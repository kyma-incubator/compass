BEGIN;

ALTER TABLE applications DROP COLUMN local_tenant_id;

COMMIT;
