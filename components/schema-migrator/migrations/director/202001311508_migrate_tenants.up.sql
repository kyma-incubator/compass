
INSERT INTO business_tenant_mappings(id)
SELECT tenant_id FROM applications a2 UNION SELECT tenant_id FROM runtimes r2 ON CONFLICT DO NOTHING;
UPDATE business_tenant_mappings SET external_tenant = id WHERE external_tenant IS NULL;
UPDATE business_tenant_mappings SET external_name = 'Tenant' WHERE external_name IS NULL;
UPDATE business_tenant_mappings SET provider_name = 'Compass' WHERE provider_name IS NULL;

ALTER TABLE api_definitions
ADD CONSTRAINT api_definitions_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE api_runtime_auths
ADD CONSTRAINT api_runtime_auths_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE applications
ADD CONSTRAINT applications_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE documents
ADD CONSTRAINT documents_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE event_api_definitions
ADD CONSTRAINT event_api_definitions_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE fetch_requests
ADD CONSTRAINT fetch_requests_tenant_id_fkey3 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE label_definitions
ADD CONSTRAINT label_definitions_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE labels
ADD CONSTRAINT labels_tenant_id_fkey2 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE runtimes
ADD CONSTRAINT runtimes_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE system_auths
ADD CONSTRAINT system_auths_tenant_id_fkey2 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE webhooks
ADD CONSTRAINT webhooks_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);