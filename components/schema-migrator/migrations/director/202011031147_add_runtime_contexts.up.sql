BEGIN;

CREATE TABLE runtime_contexts (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    runtime_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    CONSTRAINT runtime_contexts_tenant_id_fkey FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes(tenant_id, id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    CONSTRAINT runtime_contexts_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id),
    key VARCHAR(512) NOT NULL,
    value VARCHAR(512) NOT NULL
);

ALTER TABLE labels
    ADD COLUMN runtime_context_id UUID;

ALTER TABLE labels
    ADD CONSTRAINT runtime_context_id_fk FOREIGN KEY (runtime_context_id) REFERENCES runtime_contexts(id) ON DELETE CASCADE;

ALTER TABLE labels
    DROP CONSTRAINT valid_refs;

ALTER TABLE labels
    ADD CONSTRAINT valid_refs
        CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL OR labels.runtime_context_id IS NOT NULL);

DROP INDEX IF EXISTS labels_tenant_id_key_coalesce_coalesce1_idx;
CREATE UNIQUE INDEX ON labels (tenant_id, key, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), coalesce(labels.runtime_context_id, '00000000-0000-0000-0000-000000000000'));

COMMIT;
