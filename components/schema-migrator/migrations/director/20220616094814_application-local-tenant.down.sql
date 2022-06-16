BEGIN;

ALTER TABLE applications DROP COLUMN local_tenant_id VARCHAR(256);

COMMIT;
