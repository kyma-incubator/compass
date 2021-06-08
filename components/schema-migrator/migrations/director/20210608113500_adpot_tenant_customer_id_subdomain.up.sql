BEGIN;

ALTER TABLE business_tenant_mappings ADD COLUMN customer_id VARCHAR(256);
ALTER TABLE business_tenant_mappings ADD COLUMN subdomain VARCHAR(256);

COMMIT;
