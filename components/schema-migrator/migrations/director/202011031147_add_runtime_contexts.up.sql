BEGIN;

CREATE TABLE runtime_contexts (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    runtime_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes(tenant_id, id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id),
    key VARCHAR(512) NOT NULL,
    value VARCHAR(512) NOT NULL
);

COMMIT;
