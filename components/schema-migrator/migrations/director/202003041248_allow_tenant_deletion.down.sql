ALTER TABLE api_definitions
DROP CONSTRAINT api_definitions_tenant_constraint,
ADD CONSTRAINT api_definitions_tenant_id_fkey1
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE api_runtime_auths
DROP CONSTRAINT api_runtime_auths_tenant_constraint,
ADD CONSTRAINT api_runtime_auths_tenant_id_fkey
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE applications
DROP CONSTRAINT applications_tenant_constraint,
ADD CONSTRAINT applications_tenant_id_fkey
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE documents
DROP CONSTRAINT documents_tenant_constraint,
ADD CONSTRAINT documents_tenant_id_fkey1
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE event_api_definitions
DROP CONSTRAINT event_api_definitions_tenant_constraint,
ADD CONSTRAINT event_api_definitions_tenant_id_fkey1
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE fetch_requests
DROP CONSTRAINT fetch_requests_tenant_constraint,
ADD CONSTRAINT fetch_requests_tenant_id_fkey3
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE label_definitions
DROP CONSTRAINT label_definitions_tenant_constraint,
ADD CONSTRAINT label_definitions_tenant_id_fkey
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE labels
DROP CONSTRAINT labels_tenant_constraint,
ADD CONSTRAINT labels_tenant_id_fkey2
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE runtimes
DROP CONSTRAINT runtimes_tenant_constraint,
ADD CONSTRAINT runtimes_tenant_id_fkey
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE system_auths
DROP CONSTRAINT system_auths_tenant_constraint,
ADD CONSTRAINT system_auths_tenant_id_fkey2
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);

ALTER TABLE webhooks
DROP CONSTRAINT webhooks_tenant_constraint,
ADD CONSTRAINT webhooks_tenant_id_fkey1
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id);
