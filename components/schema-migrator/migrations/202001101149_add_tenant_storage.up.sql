create type tenant_status AS ENUM ('Active', 'Inactive');

create table tenant_mapping(
id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
name varchar(256),
external_tenant varchar(256) unique,
internal_tenant uuid unique,
provider_name varchar(256),
status tenant_status
);

insert into tenant_mapping(internal_id)
select tenant_id from applications a2 union select tenant_id from runtimes r2;

update tenant_mapping set external_id = internal_id;

update tenant_mapping set provider_name = 'Compass', status = 'Active';

alter table api_definitions
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table api_runtime_auths
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table applications
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table documents
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table event_api_definitions
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table fetch_requests
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table label_definitions
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table labels
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table runtimes
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table system_auths
add foreign key (tenant_id) references tenant_mapping(internal_tenant);
alter table webhooks
add foreign key (tenant_id) references tenant_mapping(internal_tenant);