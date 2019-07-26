create table label_definitions
(
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    schema JSONB
);

create unique index on label_definitions(tenant_id,key);