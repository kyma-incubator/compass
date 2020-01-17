alter table business_tenant_mappings
    drop constraint business_tenant_mappings_external_tenant_uindex;
alter table business_tenant_mappings
    add constraint business_tenant_mappings_external_tenant_provider_name_key unique (external_tenant, provider_name);
