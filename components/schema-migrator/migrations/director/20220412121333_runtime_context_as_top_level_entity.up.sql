BEGIN;

CREATE TABLE tenant_runtime_contexts
(
    tenant_id uuid NOT NULL,
    id        uuid NOT NULL,
    owner     bool,

    FOREIGN KEY (id) REFERENCES runtime_contexts (id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    PRIMARY KEY (tenant_id, id)
);

DROP VIEW IF EXISTS runtime_contexts_tenants;

COMMIT;