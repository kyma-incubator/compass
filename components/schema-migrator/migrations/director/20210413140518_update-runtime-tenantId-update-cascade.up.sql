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
COMMIT;
