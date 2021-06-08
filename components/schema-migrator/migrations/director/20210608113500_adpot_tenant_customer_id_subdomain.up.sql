BEGIN;

ALTER TABLE business_tenant_mappings ADD COLUMN customerId VARCHAR(256);
ALTER TABLE business_tenant_mappings ADD COLUMN subdomain VARCHAR(256);

COMMIT;
