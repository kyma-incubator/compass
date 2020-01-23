ALTER TABLE system_auths
    ALTER COLUMN tenant_id DROP NOT NULL;

UPDATE system_auths
    SET tenant_id = null WHERE tenant_id = '00000000-0000-0000-0000-000000000000';

ALTER TABLE system_auths ADD CONSTRAINT tenantCheck CHECK(((app_id IS NOT NULL OR runtime_id IS NOT NULL) AND tenant_id IS NOT NULL)
OR (integration_system_id IS NOT NULL AND tenant_id IS NULL));

