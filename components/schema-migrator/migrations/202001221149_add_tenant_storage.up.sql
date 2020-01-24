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

insert into business_tenant_mappings(id)
select tenant_id from applications a2 union select tenant_id from runtimes r2;

update business_tenant_mappings set external_tenant = id;

update business_tenant_mappings set external_name = 'Tenant';

update business_tenant_mappings set provider_name = 'Compass';

