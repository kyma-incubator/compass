create table label_definitions
(
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    schema JSONB
);

-- tenant_id, key unique
-- index: tenant_id, key