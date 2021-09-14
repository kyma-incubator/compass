BEGIN;

ALTER TABLE business_tenant_mappings
    DROP CONSTRAINT business_tenant_mappings_external_tenant_type_unique;
ALTER TABLE business_tenant_mappings
    ADD CONSTRAINT business_tenant_mappings_external_tenant_unique UNIQUE (external_tenant);

COMMIT;
