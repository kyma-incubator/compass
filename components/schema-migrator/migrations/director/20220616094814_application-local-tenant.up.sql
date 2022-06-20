BEGIN;

ALTER TABLE applications ADD COLUMN local_tenant_id VARCHAR(256);

COMMIT;
