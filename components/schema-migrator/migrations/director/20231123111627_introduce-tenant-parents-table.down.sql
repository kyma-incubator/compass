BEGIN;

-- TODO revert kyma-adapter webhook migration

-- Add parent column to business tenant mapping
ALTER TABLE business_tenant_mappings
    ADD COLUMN parent uuid;

-- Add business tenant mapping parent fk
ALTER TABLE business_tenant_mappings
    ADD CONSTRAINT business_tenant_mappings_parent_fk
        FOREIGN KEY (parent)
            REFERENCES business_tenant_mappings(id);

-- Create parent index
CREATE INDEX parent_index ON business_tenant_mappings(parent);

-- Fill parent column
INSERT INTO business_tenant_mappings(parent)
SELECT parent_id
    from tenant_parents join business_tenant_mappings btm on btm.id = tenant_parents.tenant_id AND btm.type IS NOT 'cost-object';

-- Drop source column
CREATE TABLE tenant_applications_temp (LIKE tenant_applications INCLUDING ALL);
-- TODO missing fk
ALTER TABLE tenant_applications_temp DROP CONSTRAINT tenant_applications_temp_pkey;
ALTER TABLE tenant_applications_temp
    ADD PRIMARY KEY (tenant_id, id);
ALTER TABLE tenant_applications_temp
    DROP COLUMN source;

INSERT INTO tenant_applications_temp(tenant_id, id, owner)
SELECT
    DISTINCT ON (id, tenant_id) tenant_id, id, owner
FROM tenant_applications
ORDER BY id, tenant_id,owner DESC;


DROP TABLE tenant_applications;
ALTER TABLE tenant_applications_temp
    RENAME TO tenant_applications; -- TODO keys and indexes are not renamed

-----------
CREATE TABLE tenant_runtimes_temp (LIKE tenant_runtimes INCLUDING ALL);

ALTER TABLE tenant_runtimes_temp DROP CONSTRAINT tenant_runtimes_temp_pkey;
ALTER TABLE tenant_runtimes_temp
    ADD PRIMARY KEY (tenant_id, id);
ALTER TABLE tenant_runtimes_temp
    DROP COLUMN source;

INSERT INTO tenant_runtimes_temp(tenant_id, id, owner)
SELECT
    DISTINCT ON (id, tenant_id) tenant_id, id, owner
FROM tenant_runtimes
ORDER BY owner DESC;


DROP TABLE tenant_runtimes;
ALTER TABLE tenant_runtimes_temp
    RENAME TO tenant_runtimes;
----------

CREATE TABLE tenant_runtime_contexts_temp (LIKE tenant_runtime_contexts INCLUDING ALL);

ALTER TABLE tenant_runtime_contexts_temp DROP CONSTRAINT tenant_runtime_contexts_temp_pkey;
ALTER TABLE tenant_runtime_contexts_temp
    ADD PRIMARY KEY (tenant_id, id);
ALTER TABLE tenant_runtime_contexts_temp
    DROP COLUMN source;

INSERT INTO tenant_runtime_contexts_temp(tenant_id, id, owner)
SELECT
    DISTINCT ON (id, tenant_id) tenant_id, id, owner
FROM tenant_runtime_contexts
ORDER BY owner DESC;


DROP TABLE tenant_runtime_contexts;
ALTER TABLE tenant_runtime_contexts_temp
    RENAME TO tenant_runtime_contexts;

-- Create tenant_id_is_direct_parent_of_target_tenant_id trigger
DROP TRIGGER tenant_id_is_direct_parent_of_target_tenant_id ON automatic_scenario_assignments;
DROP FUNCTION IF EXISTS check_tenant_id_is_direct_parent_of_target_tenant_id();

CREATE OR REPLACE FUNCTION check_tenant_id_is_direct_parent_of_target_tenant_id() RETURNS TRIGGER AS
$$
DECLARE
    count INTEGER;
BEGIN
    EXECUTE format('SELECT COUNT(1) FROM business_tenant_mappings WHERE id = %L AND parent = %L', NEW.target_tenant_id, NEW.tenant_id) INTO count;
    IF count = 0 THEN
        RAISE EXCEPTION 'target_tenant_id should be direct child of tenant_id';
    END IF;
    RETURN NULL;
END
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER tenant_id_is_direct_parent_of_target_tenant_id AFTER INSERT ON automatic_scenario_assignments
    FOR EACH ROW EXECUTE PROCEDURE check_tenant_id_is_direct_parent_of_target_tenant_id();


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
         JOIN business_tenant_mappings btm on ft.tenant_id = btm.parent;






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
         JOIN business_tenant_mappings btm on ft.tenant_id = btm.parent;

-- Drop indexes for tenant_parents table
DROP INDEX IF EXISTS tenant_parents_tenant_id;
DROP INDEX IF EXISTS tenant_parents_parent_id;

-- Drop tenant_parents table
DROP TABLE IF EXISTS tenant_parents;

COMMIT;
