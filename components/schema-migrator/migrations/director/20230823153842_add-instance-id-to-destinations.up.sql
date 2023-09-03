BEGIN;

ALTER TABLE destinations
    ADD COLUMN instance_id varchar(256);

ALTER TABLE destinations
    ADD CONSTRAINT destinations_name_tenant_instance_uniqueness UNIQUE(name, tenant_id, instance_id);

ALTER TABLE destinations
    DROP CONSTRAINT destinations_tenant_name_uniqueness;

COMMIT;
