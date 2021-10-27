BEGIN;

ALTER TABLE business_tenant_mappings
    DROP CONSTRAINT business_tenant_mappings_parent_fk,
    ADD CONSTRAINT business_tenant_mappings_parent_fk
    FOREIGN KEY (parent)
    REFERENCES business_tenant_mappings(id) ON DELETE CASCADE ;

COMMIT;
