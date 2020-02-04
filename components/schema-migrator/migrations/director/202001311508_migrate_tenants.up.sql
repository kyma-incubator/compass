
insert into business_tenant_mappings(id)
select tenant_id from applications a2 union select tenant_id from runtimes r2;
update business_tenant_mappings set external_tenant = id where external_tenant = null;
update business_tenant_mappings set external_name = 'Tenant' where external_name = null;
update business_tenant_mappings set provider_name = 'Compass' where provider_name = null;

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