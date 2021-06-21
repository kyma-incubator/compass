BEGIN;

ALTER TABLE business_tenant_mappings DROP CONSTRAINT business_tenant_mappings_parent_fk;

ALTER TABLE business_tenant_mappings DROP COLUMN parent;
ALTER TABLE business_tenant_mappings DROP COLUMN type;

DROP TYPE tenant_type;

COMMIT;
