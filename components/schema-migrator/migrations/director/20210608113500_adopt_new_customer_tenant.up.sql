BEGIN;

CREATE TYPE tenant_type AS ENUM ('account', 'customer');

ALTER TABLE business_tenant_mappings ADD COLUMN parent uuid;
ALTER TABLE business_tenant_mappings ADD COLUMN type tenant_type;

ALTER TABLE business_tenant_mappings
    ADD CONSTRAINT business_tenant_mappings_parent_fk
        FOREIGN KEY (parent)
            REFERENCES business_tenant_mappings(id);

COMMIT;
