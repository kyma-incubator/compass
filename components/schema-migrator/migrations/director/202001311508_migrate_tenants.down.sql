
alter table api_definitions
drop constraint api_definitions_tenant_id_fkey1;
alter table api_runtime_auths
drop constraint api_runtime_auths_tenant_id_fkey;
alter table applications
drop constraint applications_tenant_id_fkey;
alter table documents
drop constraint documents_tenant_id_fkey1;
alter table event_api_definitions
drop constraint event_api_definitions_tenant_id_fkey1;
alter table fetch_requests
drop constraint fetch_requests_tenant_id_fkey3;
alter table label_definitions
drop constraint label_definitions_tenant_id_fkey;
alter table labels
drop constraint labels_tenant_id_fkey2;
alter table runtimes
drop constraint runtimes_tenant_id_fkey;
alter table system_auths
drop constraint system_auths_tenant_id_fkey2;
alter table webhooks
drop constraint webhooks_tenant_id_fkey1;