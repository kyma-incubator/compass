
INSERT INTO business_tenant_mappings(id)
SELECT tenant_id FROM applications a2 UNION SELECT tenant_id FROM runtimes r2 ON CONFLICT DO NOTHING;
UPDATE business_tenant_mappings SET external_tenant = id WHERE external_tenant IS NULL;
UPDATE business_tenant_mappings SET external_name = 'Tenant' WHERE external_name IS NULL;
UPDATE business_tenant_mappings SET provider_name = 'Compass' WHERE provider_name IS NULL;

ALTER TABLE api_definitions
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE api_runtime_auths
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE applications
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE documents
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE event_api_definitions
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE fetch_requests
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE label_definitions
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE labels
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE runtimes
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE system_auths
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);
ALTER TABLE webhooks
ADD FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id);