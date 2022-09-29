BEGIN;

DROP TABLE formation_assignments;

ALTER TABLE formations
    DROP CONSTRAINT formation_id_tenant_id_unique;

COMMIT;
