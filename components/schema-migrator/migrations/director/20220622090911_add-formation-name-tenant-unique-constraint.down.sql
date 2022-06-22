BEGIN;

ALTER TABLE formations
    DROP CONSTRAINT formation_name_tenant_id_unique;

COMMIT;
