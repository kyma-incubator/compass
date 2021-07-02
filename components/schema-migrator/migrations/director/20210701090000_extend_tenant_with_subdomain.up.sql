BEGIN;

ALTER TABLE business_tenant_mappings ADD COLUMN subdomain VARCHAR(255);

COMMIT;
