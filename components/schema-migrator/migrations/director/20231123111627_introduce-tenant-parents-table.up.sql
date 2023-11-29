BEGIN;

-- Create many to many tenant_parents table
CREATE TABLE tenant_parents
(
    tenant_id   uuid NOT NULL,
    parent_id   uuid NOT NULL,

    CONSTRAINT tenant_parents_tenant_id_fkey  FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    CONSTRAINT tenant_parents_parent_id_fkey  FOREIGN KEY (parent_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    PRIMARY KEY (tenant_id, parent_id)
);

-- Create indexes for tenant_parents table
CREATE INDEX tenant_parents_tenant_id ON tenant_parents(tenant_id);
CREATE INDEX tenant_parents_parent_id ON tenant_parents(parent_id);


-- Populate 'tenant_parents' table with data from 'business_tenant_mappings'
INSERT INTO tenant_parents (tenant_id, parent_id)
SELECT id, parent
FROM business_tenant_mappings
WHERE parent IS NOT NULL;


-- TODO make source column not null
-- Add source column to tenant_applications table
ALTER TABLE tenant_applications
    ADD COLUMN source uuid;
ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_applications
    ADD PRIMARY KEY (tenant_id, id, source);

-- Add source column to tenant_runtimes table
ALTER TABLE tenant_runtimes
    ADD COLUMN source uuid;
ALTER TABLE tenant_runtimes
    ADD CONSTRAINT tenant_runtimes_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_runtimes
    ADD PRIMARY KEY (tenant_id, id, source);

-- Add source column to tenant_runtime_contexts table
ALTER TABLE tenant_runtime_contexts
    ADD COLUMN source uuid;
ALTER TABLE tenant_runtime_contexts
    ADD CONSTRAINT tenant_runtime_contexts_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_runtime_contexts
    ADD PRIMARY KEY (tenant_id, id, source);


DROP VIEW IF EXISTS formation_templates_webhooks_tenants;

CREATE OR REPLACE VIEW formation_templates_webhooks_tenants (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
                                                             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
                                                             app_template_id, formation_template_id, tenant_id, owner)
AS
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       w.formation_template_id,
       ft.tenant_id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id
UNION ALL
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       w.formation_template_id,
       btm.id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id
         JOIN tenant_parents tp on ft.tenant_id = tp.parent_id;





DROP VIEW IF EXISTS webhooks_tenants;

CREATE OR REPLACE VIEW webhooks_tenants
            (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
             input_template, header_template, output_template, status_template, runtime_id, integration_system_id,
             app_template_id, formation_template_id, tenant_id, owner)
AS
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       w.formation_template_id,
       ta.tenant_id,
       ta.owner
FROM webhooks w
         JOIN tenant_applications ta ON w.app_id = ta.id
UNION ALL
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       w.formation_template_id,
       tr.tenant_id,
       tr.owner
FROM webhooks w
         JOIN tenant_runtimes tr ON w.runtime_id = tr.id
UNION ALL
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       w.formation_template_id,
       ft.tenant_id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id
UNION ALL
SELECT w.id,
       w.app_id,
       w.url,
       w.type,
       w.auth,
       w.mode,
       w.correlation_id_key,
       w.retry_interval,
       w.timeout,
       w.url_template,
       w.input_template,
       w.header_template,
       w.output_template,
       w.status_template,
       w.runtime_id,
       w.integration_system_id,
       w.app_template_id,
       w.formation_template_id,
       btm.id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id
         JOIN tenant_parents tp on ft.tenant_id = tp.parent_id;

-- Drop 'parent' column from 'business_tenant_mappings'
ALTER TABLE business_tenant_mappings
DROP COLUMN parent;

COMMIT;
