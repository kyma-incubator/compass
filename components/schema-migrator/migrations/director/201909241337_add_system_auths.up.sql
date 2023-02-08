CREATE TABLE system_auths (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    app_id uuid,
    constraint system_auths_tenant_id_fkey foreign key (tenant_id, app_id) references applications (tenant_id, id) ON DELETE CASCADE,
    runtime_id uuid,
    constraint system_auths_tenant_id_fkey1 foreign key (tenant_id, runtime_id) references runtimes (tenant_id, id) ON DELETE CASCADE,
    integration_system_id uuid,
    value jsonb,
    CONSTRAINT valid_refs CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL OR integration_system_id IS NOT NULL)
);

CREATE INDEX ON system_auths (tenant_id);
CREATE UNIQUE INDEX ON system_auths (id, tenant_id);
