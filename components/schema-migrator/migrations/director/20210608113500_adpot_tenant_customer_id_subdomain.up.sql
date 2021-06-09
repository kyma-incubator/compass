BEGIN;

ALTER TABLE business_tenant_mappings ADD COLUMN customer_id VARCHAR(256);
ALTER TABLE business_tenant_mappings ADD COLUMN subdomain VARCHAR(256);

ALTER TABLE business_tenant_mappings DROP CONSTRAINT business_tenant_mappings_external_tenant_unique;
ALTER TABLE business_tenant_mappings ADD CONSTRAINT business_tenant_mappings_external_tenant_customer_id_unique UNIQUE (external_tenant, customer_id);

COMMIT;
