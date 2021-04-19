BEGIN;
ALTER TABLE labels DROP CONSTRAINT labels_tenant_id_fkey1;
ALTER TABLE labels
    ADD CONSTRAINT labels_tenant_id_fkey1
        FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;

ALTER TABLE system_auths DROP CONSTRAINT system_auths_tenant_id_fkey1;
ALTER TABLE system_auths
    ADD CONSTRAINT system_auths_tenant_id_fkey1
        FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;

ALTER TABLE runtime_contexts DROP CONSTRAINT runtime_contexts_tenant_id_fkey;
ALTER TABLE runtime_contexts
    ADD CONSTRAINT runtime_contexts_tenant_id_fkey
        FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;

ALTER TABLE webhooks DROP CONSTRAINT webhooks_runtime_id_fkey;
ALTER TABLE webhooks
    ADD CONSTRAINT webhooks_runtime_id_fkey
        FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;

ALTER TABLE api_runtime_auths DROP CONSTRAINT runtime_auths_tenant_id_fkey;
ALTER TABLE api_runtime_auths
    ADD CONSTRAINT runtime_auths_tenant_id_fkey
        FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;

COMMIT;
