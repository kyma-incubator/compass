BEGIN;

ALTER TABLE formations
    ADD CONSTRAINT formation_name_tenant_id_unique UNIQUE (name, tenant_id);

COMMIT;
