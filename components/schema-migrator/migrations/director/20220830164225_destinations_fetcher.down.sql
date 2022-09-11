BEGIN;
DROP VIEW tenants_destinations;

ALTER TABLE destinations
    DROP CONSTRAINT destinations_tenant_name_uniqueness;

DROP TABLE destinations;
COMMIT;
