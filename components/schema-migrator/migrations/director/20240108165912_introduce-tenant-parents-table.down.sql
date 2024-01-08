BEGIN;

UPDATE webhooks
SET input_template = '{"context":{"platform":"{{if .CustomerTenantContext.AccountID}}btp{{else}}unified-services{{end}}","uclFormationId":"{{.FormationID}}","accountId":"{{if .CustomerTenantContext.AccountID}}{{.CustomerTenantContext.AccountID}}{{else}}{{.CustomerTenantContext.Path}}{{end}}","crmId":"{{.CustomerTenantContext.CustomerID}}","operation":"{{.Operation}}"},"assignedTenant":{"state":"{{.Assignment.State}}","uclAssignmentId":"{{.Assignment.ID}}","deploymentRegion":"{{if .Application.Labels.region}}{{.Application.Labels.region}}{{else}}{{.ApplicationTemplate.Labels.region}}{{end}}","applicationNamespace":"{{if .Application.ApplicationNamespace}}{{.Application.ApplicationNamespace}}{{else}}{{.ApplicationTemplate.ApplicationNamespace}}{{end}}","applicationUrl":"{{.Application.BaseURL}}","applicationTenantId":"{{.Application.LocalTenantID}}","uclSystemName":"{{.Application.Name}}","uclSystemTenantId":"{{.Application.ID}}",{{if .ApplicationTemplate.Labels.parameters}}"parameters":{{.ApplicationTemplate.Labels.parameters}},{{end}}"configuration":{{.ReverseAssignment.Value}}},"receiverTenant":{"ownerTenants": [{{ Join .Runtime.Tenant.Parents }}],"state":"{{.ReverseAssignment.State}}","uclAssignmentId":"{{.ReverseAssignment.ID}}","deploymentRegion":"{{if and .RuntimeContext .RuntimeContext.Labels.region}}{{.RuntimeContext.Labels.region}}{{else}}{{.Runtime.Labels.region}}{{end}}","applicationNamespace":"{{.Runtime.ApplicationNamespace}}","applicationTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.Value}}{{else}}{{.Runtime.Labels.global_subaccount_id}}{{end}}","uclSystemTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}",{{if .Runtime.Labels.parameters}}"parameters":{{.Runtime.Labels.parameters}},{{end}}"configuration":{{.Assignment.Value}}}}'
where runtime_id IN
      (SELECT runtime_id
       FROM labels
       WHERE runtime_id IS NOT NULL
         AND
    key = 'runtimeType'
  AND value = '"kyma"'
    );

-- Add parent column to business tenant mapping
ALTER TABLE business_tenant_mappings
    ADD COLUMN parent uuid;

-- Add business tenant mapping parent fk
ALTER TABLE business_tenant_mappings
    ADD CONSTRAINT business_tenant_mappings_parent_fk
        FOREIGN KEY (parent)
            REFERENCES business_tenant_mappings (id);

-- Create parent index
CREATE INDEX parent_index ON business_tenant_mappings (parent);

-- Fill parent column
UPDATE business_tenant_mappings SET parent=parent_id
FROM tenant_parents
WHERE  business_tenant_mappings.id = tenant_parents.tenant_id AND business_tenant_mappings.type <> 'cost-object'::tenant_type;

-- tenant_applications
DELETE
FROM tenant_applications t1 USING tenant_applications t2
WHERE t1.source
    < t2.source
  AND t1.tenant_id = t2.tenant_id
  AND t1.id=t2.id
  AND t1.owner=t2.owner;

DELETE
FROM tenant_applications t1 USING tenant_applications t2
WHERE t1.tenant_id = t2.tenant_id AND t1.id=t2.id AND t1.owner= false AND t2.owner= true;

ALTER TABLE tenant_applications DROP column source;

-- tenant_runtimes
DELETE
FROM tenant_runtimes t1 USING tenant_runtimes t2
WHERE t1.source
    < t2.source
  AND t1.tenant_id = t2.tenant_id
  AND t1.id=t2.id
  AND t1.owner=t2.owner;

DELETE
FROM tenant_runtimes t1 USING tenant_runtimes t2
WHERE t1.tenant_id = t2.tenant_id AND t1.id=t2.id AND t1.owner= false AND t2.owner= true;

ALTER TABLE tenant_runtimes DROP column source;

-- tenant_runtime_contexts
DELETE
FROM tenant_runtime_contexts t1 USING tenant_runtime_contexts t2
WHERE t1.source
    < t2.source
  AND t1.tenant_id = t2.tenant_id
  AND t1.id=t2.id
  AND t1.owner=t2.owner;

DELETE
FROM tenant_runtime_contexts t1 USING tenant_runtime_contexts t2
WHERE t1.tenant_id = t2.tenant_id AND t1.id=t2.id AND t1.owner= false AND t2.owner= true;

ALTER TABLE tenant_runtime_contexts DROP column source;

-- Identify duplicates and keep the one with owner=true
-- WITH ranked_rows AS (
--     SELECT
--         tenant_id,
--         id,
--         source,
--         ROW_NUMBER() OVER (PARTITION BY tenant_id, id ORDER BY owner DESC) AS row_num
--     FROM
--         tenant_applications
-- )
-- DELETE FROM tenant_applications
-- WHERE (tenant_id, id, source) IN (SELECT tenant_id, id, source FROM ranked_rows WHERE row_num > 1);

-- Create tenant_id_is_direct_parent_of_target_tenant_id trigger
DROP TRIGGER tenant_id_is_direct_parent_of_target_tenant_id ON automatic_scenario_assignments;
DROP FUNCTION IF EXISTS check_tenant_id_is_direct_parent_of_target_tenant_id();

CREATE
OR REPLACE FUNCTION check_tenant_id_is_direct_parent_of_target_tenant_id() RETURNS TRIGGER AS
$$
DECLARE
count INTEGER;
BEGIN
EXECUTE format('SELECT COUNT(1) FROM business_tenant_mappings WHERE id = %L AND parent = %L', NEW.target_tenant_id,
               NEW.tenant_id) INTO count;
IF
count = 0 THEN
        RAISE EXCEPTION 'target_tenant_id should be direct child of tenant_id';
END IF;
RETURN NULL;
END
$$
LANGUAGE plpgsql;

CREATE
CONSTRAINT TRIGGER tenant_id_is_direct_parent_of_target_tenant_id AFTER INSERT ON automatic_scenario_assignments
    FOR EACH ROW EXECUTE PROCEDURE check_tenant_id_is_direct_parent_of_target_tenant_id();


DROP VIEW IF EXISTS formation_templates_webhooks_tenants;


CREATE
OR REPLACE VIEW formation_templates_webhooks_tenants (id, app_id, url, type, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
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
CREATE
OR REPLACE VIEW webhooks_tenants
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
