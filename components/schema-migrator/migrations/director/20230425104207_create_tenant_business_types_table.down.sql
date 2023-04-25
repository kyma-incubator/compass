BEGIN;

ALTER TABLE applications DROP COLUMN IF EXISTS tenant_business_type_id;

ALTER TABLE applications
    DROP CONSTRAINT IF EXISTS applications_tenant_business_type_id_fk;

DROP TABLE IF EXISTS tenant_business_types;

COMMIT;