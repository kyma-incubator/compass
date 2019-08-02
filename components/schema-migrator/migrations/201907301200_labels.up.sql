create table labels
(
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    app_id     varchar(100), -- TODO: Change to Application UUID reference
    runtime_id uuid REFERENCES runtimes (id) ON DELETE CASCADE,
    value      JSONB
);

create unique index on labels (tenant_id, key, runtime_id, app_id);