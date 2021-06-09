BEGIN;

ALTER TABLE business_tenant_mappings DROP CONSTRAINT business_tenant_mappings_external_tenant_customer_id_unique;
ALTER TABLE business_tenant_mappings ADD CONSTRAINT business_tenant_mappings_external_tenant_unique UNIQUE (external_tenant);

ALTER TABLE business_tenant_mappings DROP COLUMN customer_id;
ALTER TABLE business_tenant_mappings DROP COLUMN subdomain;

COMMIT;
