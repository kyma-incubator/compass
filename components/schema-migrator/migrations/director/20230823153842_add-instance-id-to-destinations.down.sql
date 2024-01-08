BEGIN;

ALTER TABLE destinations
    ADD CONSTRAINT destinations_tenant_name_uniqueness UNIQUE(name, tenant_id);

ALTER TABLE destinations
    DROP CONSTRAINT destinations_name_tenant_instance_uniqueness;

ALTER TABLE destinations
    DROP COLUMN instance_id;

COMMIT;
