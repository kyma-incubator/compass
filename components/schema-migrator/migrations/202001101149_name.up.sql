create table tenant_mapping(
name varchar(256),
external_id uuid unique,
internal_id uuid unique,
provider varchar(256)
);

insert into tenant_mapping(internal_id)
select tenant_id from api_definitions ad union select tenant_id from api_runtime_auths ara
union select tenant_id from applications a2 union select tenant_id from documents d2
union select tenant_id from event_api_definitions ead union select tenant_id from fetch_requests fr2
union select tenant_id from label_definitions ld union select tenant_id from runtimes r2
union select tenant_id from system_auths sa union select tenant_id from webhooks w2;

update tenant_mapping set external_id = internal_id;

update tenant_mapping set provider = 'Compass';

alter table api_definitions
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table api_runtime_auths
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table applications
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table documents
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table event_api_definitions
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table fetch_requests
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table label_definitions
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table labels
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table runtimes
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table system_auths
add foreign key (tenant_id) references tenant_mapping(internal_id);
alter table webhooks
add foreign key (tenant_id) references tenant_mapping(internal_id);