UPDATE system_auths SET tenant_id='00000000-0000-0000-0000-000000000000' WHERE tenant_id IS NULL;

ALTER TABLE system_auths
    DROP CONSTRAINT tenantCheck;

ALTER TABLE system_auths
    ALTER COLUMN tenant_id SET NOT NULL;
