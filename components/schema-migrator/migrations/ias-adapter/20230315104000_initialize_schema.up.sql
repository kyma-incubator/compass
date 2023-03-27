CREATE TABLE IF NOT EXISTS tenant_mappings (
    formation_id uuid,
    ucl_application_id uuid,
    value jsonb NOT NULL,
    CONSTRAINT pk PRIMARY KEY(formation_id, ucl_application_id)
);