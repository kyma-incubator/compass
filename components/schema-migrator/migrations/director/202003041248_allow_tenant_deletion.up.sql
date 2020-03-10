ALTER TABLE api_definitions
DROP CONSTRAINT api_definitions_tenant_id_fkey1,
ADD CONSTRAINT api_definitions_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE api_runtime_auths
DROP CONSTRAINT api_runtime_auths_tenant_id_fkey,
ADD CONSTRAINT api_runtime_auths_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;



ALTER TABLE applications
DROP CONSTRAINT applications_tenant_id_fkey,
ADD CONSTRAINT applications_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE documents
DROP CONSTRAINT documents_tenant_id_fkey1,
ADD CONSTRAINT documents_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE event_api_definitions
DROP CONSTRAINT event_api_definitions_tenant_id_fkey1,
ADD CONSTRAINT event_api_definitions_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE fetch_requests
DROP CONSTRAINT fetch_requests_tenant_id_fkey3,
ADD CONSTRAINT fetch_requests_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE label_definitions
DROP CONSTRAINT label_definitions_tenant_id_fkey,
ADD CONSTRAINT label_definitions_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE labels
DROP CONSTRAINT labels_tenant_id_fkey2,
ADD CONSTRAINT labels_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE runtimes
DROP CONSTRAINT runtimes_tenant_id_fkey,
ADD CONSTRAINT runtimes_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE system_auths
DROP CONSTRAINT system_auths_tenant_id_fkey2,
ADD CONSTRAINT system_auths_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;

ALTER TABLE webhooks
DROP CONSTRAINT webhooks_tenant_id_fkey1,
ADD CONSTRAINT webhooks_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;
