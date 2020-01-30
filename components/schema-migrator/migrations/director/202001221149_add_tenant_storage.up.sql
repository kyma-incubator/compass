create type tenant_status AS ENUM ('Active', 'Inactive');

create table business_tenant_mappings(
id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
external_name varchar(256),
external_tenant varchar(256),
provider_name varchar(256),
status tenant_status default 'Active'
);

alter table business_tenant_mappings
add unique (external_tenant, provider_name);
