BEGIN;

ALTER TABLE business_tenant_mappings DROP COLUMN subdomain;

COMMIT;
