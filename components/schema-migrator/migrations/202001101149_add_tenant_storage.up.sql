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

insert into business_tenant_mappings(id, external_name, external_tenant, provider_name, status) values('3e64ebae-38b5-46a0-b1ed-9ccee153a0ae','Default',
'3e64ebae-38b5-46a0-b1ed-9ccee153a0ae','Compass','Active');


insert into business_tenant_mappings(id, external_name, external_tenant, provider_name, status) values('2a1502ba-aded-11e9-a2a3-2a2ae2dbcce4','Default for tests',
'2a1502ba-aded-11e9-a2a3-2a2ae2dbcce4','Compass','Active');


insert into business_tenant_mappings(id, external_name, external_tenant, provider_name, status) values('1eba80dd-8ff6-54ee-be4d-77944d17b10b','Default for tests',
'1eba80dd-8ff6-54ee-be4d-77944d17b10b','Compass','Active');


insert into business_tenant_mappings(id, external_name, external_tenant, provider_name, status) values('9ca034f1-11ab-5b25-b76f-dc77106f571d','Default for tests',
'9ca034f1-11ab-5b25-b76f-dc77106f571d','Compass','Active');


alter table api_definitions
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table api_runtime_auths
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table applications
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table documents
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table event_api_definitions
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table fetch_requests
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table label_definitions
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table labels
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table runtimes
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table system_auths
add foreign key (tenant_id) references business_tenant_mappings(id);
alter table webhooks
add foreign key (tenant_id) references business_tenant_mappings(id);