create table labels
(
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    label_def_id uuid REFERENCES label_definitions(id) NOT NULL,
    app_id     varchar(100), -- TODO: Change to Application UUID reference
    runtime_id uuid REFERENCES runtimes (id),
    value      JSONB
);

create unique index on labels (tenant_id, key);