alter table business_tenant_mappings
    drop constraint business_tenant_mappings_external_tenant_provider_name_key;
alter table business_tenant_mappings
    add constraint business_tenant_mappings_external_tenant_uindex unique (external_tenant);