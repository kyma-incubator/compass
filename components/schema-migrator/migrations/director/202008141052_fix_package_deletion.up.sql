BEGIN;

ALTER TABLE packages
DROP CONSTRAINT packages_tenant_id_fkey1,
ADD CONSTRAINT packages_tenant_id_fkey1
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE package_instance_auths
DROP CONSTRAINT package_instance_auths_tenant_id_fkey1,
ADD CONSTRAINT package_instance_auths_tenant_id_fkey1
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

COMMIT;
