BEGIN;

ALTER TABLE formations
    ADD CONSTRAINT formation_id_tenant_id_unique UNIQUE (id, tenant_id);

CREATE TABLE formation_assignments (
    id           UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    formation_id UUID NOT NULL CHECK    (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id    UUID NOT NULL CHECK    (id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (formation_id, tenant_id) REFERENCES formations(id, tenant_id) ON DELETE CASCADE,
    source       VARCHAR(256) NOT NULL,
    source_type  VARCHAR(256) NOT NULL,
    target       VARCHAR(256) NOT NULL,
    target_type  VARCHAR(256) NOT NULL,
    state        VARCHAR(256) NOT NULL,
    value        JSONB,
    CONSTRAINT formation_assignments_formation_id_source_target_unique UNIQUE (formation_id, source, target)
);

CREATE INDEX idx_formation_assignments_formation_id
    ON formation_assignments (formation_id);

CREATE INDEX idx_formation_assignments_tenant_id
    ON formation_assignments (tenant_id);

COMMIT;
