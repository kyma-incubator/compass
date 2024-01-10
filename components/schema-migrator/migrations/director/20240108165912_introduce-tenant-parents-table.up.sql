BEGIN;

UPDATE webhooks
SET input_template = '{"context":{"platform":"{{if .CustomerTenantContext.AccountID}}btp{{else}}unified-services{{end}}","uclFormationId":"{{.FormationID}}","accountId":"{{if .CustomerTenantContext.AccountID}}{{.CustomerTenantContext.AccountID}}{{else}}{{.CustomerTenantContext.Path}}{{end}}","crmId":"{{.CustomerTenantContext.CustomerID}}","operation":"{{.Operation}}"},"assignedTenant":{"state":"{{.Assignment.State}}","uclAssignmentId":"{{.Assignment.ID}}","deploymentRegion":"{{if .Application.Labels.region}}{{.Application.Labels.region}}{{else}}{{.ApplicationTemplate.Labels.region}}{{end}}","applicationNamespace":"{{if .Application.ApplicationNamespace}}{{.Application.ApplicationNamespace}}{{else}}{{.ApplicationTemplate.ApplicationNamespace}}{{end}}","applicationUrl":"{{.Application.BaseURL}}","applicationTenantId":"{{.Application.LocalTenantID}}","uclSystemName":"{{.Application.Name}}","uclSystemTenantId":"{{.Application.ID}}",{{if .ApplicationTemplate.Labels.parameters}}"parameters":{{.ApplicationTemplate.Labels.parameters}},{{end}}"configuration":{{.ReverseAssignment.Value}}},"receiverTenant":{"ownerTenant":"{{.Runtime.Tenant.Parent}}","state":"{{.ReverseAssignment.State}}","uclAssignmentId":"{{.ReverseAssignment.ID}}","deploymentRegion":"{{if and .RuntimeContext .RuntimeContext.Labels.region}}{{.RuntimeContext.Labels.region}}{{else}}{{.Runtime.Labels.region}}{{end}}","applicationNamespace":"{{.Runtime.ApplicationNamespace}}","applicationTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.Value}}{{else}}{{.Runtime.Labels.global_subaccount_id}}{{end}}","uclSystemTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}",{{if .Runtime.Labels.parameters}}"parameters":{{.Runtime.Labels.parameters}},{{end}}"configuration":{{.Assignment.Value}}}}'
where runtime_id IN
      (SELECT runtime_id
       FROM labels
       WHERE runtime_id IS NOT NULL
         AND
    key = 'runtimeType'
  AND value = '"kyma"'
    );

-- Create many to many tenant_parents table
CREATE TABLE tenant_parents
(
    tenant_id uuid NOT NULL,
    parent_id uuid NOT NULL,

    CONSTRAINT tenant_parents_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    CONSTRAINT tenant_parents_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    PRIMARY KEY (tenant_id, parent_id)
);

-- Copy business_tenant_mappings table
CREATE TABLE business_tenant_mappings_temp AS TABLE business_tenant_mappings;
CREATE INDEX business_tenant_mappings_temp_external_tenant ON business_tenant_mappings_temp (external_tenant);
CREATE INDEX business_tenant_mappings_temp_id ON business_tenant_mappings_temp (id);

-- Populate 'tenant_parents' table with data from 'business_tenant_mappings'
INSERT INTO tenant_parents (tenant_id, parent_id)
SELECT id, parent
FROM business_tenant_mappings_temp
WHERE parent IS NOT NULL;

-- Create indexes for tenant_parents table
CREATE INDEX tenant_parents_tenant_id ON tenant_parents (tenant_id);
CREATE INDEX tenant_parents_parent_id ON tenant_parents (parent_id);

-- Migrate tenant access records for applications

--
CREATE TABLE tenant_applications_temp AS TABLE tenant_applications;

CREATE INDEX tenant_applications_temp_tenant_id ON tenant_applications_temp (tenant_id);
CREATE INDEX tenant_applications_temp_app_id ON tenant_applications_temp (id);


-- Add source column to tenant_applications_temp table
ALTER TABLE tenant_applications_temp
    ADD COLUMN source uuid;

-- ALTER TABLE tenant_applications_temp DROP CONSTRAINT tenant_applications_temp_pkey;
ALTER TABLE  tenant_applications_temp
    ADD CONSTRAINT unique_tenant_applications_temp UNIQUE (tenant_id,id,source);

-- Add tenant access for the parents of the owning tenant
INSERT INTO tenant_applications_temp
    (select tp.parent_id, ta.id, ta.owner, ta.tenant_id
     from tenant_applications_temp ta
               JOIN tenant_parents tp on ta.tenant_id = tp.tenant_id) ON CONFLICT (tenant_id,id,source) DO NOTHING ;

-- Add tenant access records for the owning tenant itself
update tenant_applications_temp ta SET source = ta.tenant_id
where ta.source is null and NOT EXISTS (select 1 from tenant_applications_temp parent_ta
                                                          join tenant_parents tp on parent_ta.tenant_id=tp.parent_id
                                                          join tenant_applications_temp child_ta on child_ta.tenant_id = tp.tenant_id and parent_ta.ID = child_ta.ID
                                        where ta.tenant_id=parent_ta.tenant_id AND ta.id =parent_ta.id);

-- Delete tenant access records with null source as they leftover access records for the parents of the owning tenant that were already populated by the previous queries
delete from tenant_applications_temp where source is null;

-- Tenant access records that originated from the directive were processed in the reverse direction. Now according to the TA table the CRM knows about the resource from the GA
-- We have to swap the source and tenant_id for records where the source is GA, the tenant_id is CRM and there is a record for the same resource where the tenant_id is resource-group (this means that the resource was created in Atom)
update tenant_applications_temp ta
  set source = ta.tenant_id,  tenant_id = ta.source
  from tenant_applications_temp ta1
         join business_tenant_mappings_temp btm on btm.id = ta1.tenant_id AND btm.type = 'customer'::tenant_type
         join business_tenant_mappings_temp btm2 on btm2.id = ta1.source AND btm2.type = 'account'::tenant_type
         join tenant_applications_temp ta2 on ta1.id=ta2.id
         join business_tenant_mappings_temp btm3 on btm3.id = ta2.tenant_id AND btm3.type = 'resource-group'::tenant_type
 where ta.tenant_id=ta1.tenant_id and ta.source=ta1.source and ta.id=ta1.id and ta.owner = ta1.owner;

-- After fixing the swapped source and tenant_id there are leftover records stating that the GA knows about the resource from itself which is not correct as The GA knows about the resource from the CRM. Such records should be deleted
delete from tenant_applications_temp ta
using
        tenant_applications_temp ta2
       join tenant_parents tp on tp.parent_id = ta2.source and tp.tenant_id=ta2.tenant_id
    where ta.tenant_id =ta2.tenant_id AND ta.id=ta2.id and ta.tenant_id = ta.source;

ALTER TABLE tenant_applications RENAME TO tenant_applications_old;
ALTER TABLE tenant_applications_old RENAME CONSTRAINT tenant_applications_pkey TO tenant_applications_old_pkey;
ALTER TABLE tenant_applications_old RENAME CONSTRAINT tenant_applications_id_fkey TO tenant_applications_old_id_fkey;
ALTER TABLE tenant_applications_old RENAME CONSTRAINT tenant_applications_tenant_id_fkey TO tenant_applications_old_tenant_id_fkey;
ALTER INDEX tenant_applications_app_id RENAME TO tenant_applications_old_app_id;
ALTER INDEX tenant_applications_tenant_id RENAME TO tenant_applications_old_tenant_id;

ALTER TABLE tenant_applications_temp RENAME TO tenant_applications;
ALTER TABLE tenant_applications
    ADD PRIMARY KEY (tenant_id, id, source);
ALTER TABLE tenant_applications DROP CONSTRAINT unique_tenant_applications_temp;


ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_tenant_id_fk
        FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_id_fk
        FOREIGN KEY (id) REFERENCES applications (id) ON DELETE CASCADE;
ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;

ALTER TABLE tenant_applications
    alter column source set not null;

ALTER INDEX tenant_applications_temp_app_id RENAME TO tenant_applications_app_id;
ALTER INDEX tenant_applications_temp_tenant_id RENAME TO tenant_applications_tenant_id;
CREATE INDEX tenant_applications_source ON tenant_applications (source);

-- Migrate tenant access records for runtimes

-- Add source column to tenant_runtimes table
ALTER TABLE tenant_runtimes
    ADD COLUMN source uuid;
ALTER TABLE tenant_runtimes
    ADD CONSTRAINT tenant_runtimes_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_runtimes DROP CONSTRAINT tenant_runtimes_pkey;

ALTER TABLE  tenant_runtimes
    ADD CONSTRAINT unique_tenant_runtimes UNIQUE (tenant_id,id,source);

-- Add tenant access for the parents of the owning tenant
INSERT INTO tenant_runtimes
    (SELECT tp.parent_id, ta.id, ta.owner, ta.tenant_id
     FROM tenant_runtimes ta
              JOIN tenant_parents tp ON ta.tenant_id = tp.tenant_id)ON CONFLICT (tenant_id,id,source) DO NOTHING ;

-- Add tenant access records for the owning tenant itself
UPDATE tenant_runtimes ta SET source = ta.tenant_id
WHERE ta.source IS NULL AND NOT EXISTS (SELECT 1 FROM tenant_runtimes parent_ta
                                                          JOIN tenant_parents tp ON parent_ta.tenant_id=tp.parent_id
                                                          JOIN tenant_runtimes child_ta ON child_ta.tenant_id = tp.tenant_id AND parent_ta.ID = child_ta.ID
                                        WHERE ta.tenant_id=parent_ta.tenant_id AND ta.id =parent_ta.id);

-- Delete tenant access records with null source as they leftover access records for the parents of the owning tenant that were already populated by the previous queries
DELETE FROM tenant_runtimes WHERE source IS NULL;

-- Tenant access records that originated from the directive were processed in the reverse direction. Now according to the TA table the CRM knows about the resource from the GA
-- We have to swap the source and tenant_id for records where the source is GA, the tenant_id is CRM and there is a record for the same resource where the tenant_id is resource-group (this means that the resource was created in Atom)
UPDATE tenant_runtimes ta
SET source = ta.tenant_id,  tenant_id = ta.source
FROM tenant_runtimes ta1
         JOIN business_tenant_mappings btm ON btm.id = ta1.tenant_id AND btm.type = 'customer'::tenant_type
         JOIN business_tenant_mappings btm2 ON btm2.id = ta1.source AND btm2.type = 'account'::tenant_type
         JOIN tenant_runtimes ta2 ON ta1.id=ta2.id
         JOIN business_tenant_mappings btm3 ON btm3.id = ta2.tenant_id AND btm3.type = 'resource-group'::tenant_type
WHERE ta.tenant_id=ta1.tenant_id AND ta.source=ta1.source AND ta.id=ta1.id AND ta.owner = ta1.owner;

-- After fixing the swapped source and tenant_id there are leftover records stating that the GA knows about the resource from itself which is not correct as The GA knows about the resource from the CRM. Such records should be deleted
DELETE FROM tenant_runtimes ta
    USING
        tenant_runtimes ta2
            JOIN tenant_parents tp ON tp.parent_id = ta2.source AND tp.tenant_id=ta2.tenant_id
WHERE ta.tenant_id =ta2.tenant_id AND ta.id=ta2.id AND ta.tenant_id = ta.source;

ALTER TABLE tenant_runtimes
    ALTER COLUMN source SET NOT NULL;

ALTER TABLE tenant_runtimes
    ADD PRIMARY KEY (tenant_id, id, source);

CREATE INDEX tenant_runtimes_source ON tenant_runtimes (source);


-- Migrate tenant access records for runtime_contexts

-- Add source column to tenant_runtime_contexts table
ALTER TABLE tenant_runtime_contexts
    ADD COLUMN source uuid;
ALTER TABLE tenant_runtime_contexts
    ADD CONSTRAINT tenant_runtime_contexts_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_runtime_contexts DROP CONSTRAINT tenant_runtime_contexts_pkey;

ALTER TABLE  tenant_runtime_contexts
    ADD CONSTRAINT unique_tenant_runtime_contexts UNIQUE (tenant_id,id,source);

-- Add tenant access for the parents of the owning tenant
INSERT INTO tenant_runtime_contexts
    (SELECT tp.parent_id, ta.id, ta.owner, ta.tenant_id
     FROM tenant_runtime_contexts ta
              JOIN tenant_parents tp ON ta.tenant_id = tp.tenant_id)ON CONFLICT (tenant_id,id,source) DO NOTHING ;

-- Add tenant access records for the owning tenant itself
UPDATE tenant_runtime_contexts ta SET source = ta.tenant_id
WHERE ta.source is null and NOT EXISTS (SELECT 1 FROM tenant_runtime_contexts parent_ta
                                                          JOIN tenant_parents tp on parent_ta.tenant_id=tp.parent_id
                                                          JOIN tenant_runtime_contexts child_ta on child_ta.tenant_id = tp.tenant_id and parent_ta.ID = child_ta.ID
                                        where ta.tenant_id=parent_ta.tenant_id AND ta.id =parent_ta.id);

-- Delete tenant access records with null source as they leftover access records for the parents of the owning tenant that were already populated by the previous queries
delete from tenant_runtime_contexts where source is null;

-- Tenant access records that originated from the directive were processed in the reverse direction. Now according to the TA table the CRM knows about the resource from the GA
-- We have to swap the source and tenant_id for records where the source is GA, the tenant_id is CRM and there is a record for the same resource where the tenant_id is resource-group (this means that the resource was created in Atom)
UPDATE tenant_runtime_contexts ta
SET source = ta.tenant_id,  tenant_id = ta.source
FROM tenant_runtime_contexts ta1
         JOIN business_tenant_mappings btm ON btm.id = ta1.tenant_id AND btm.type = 'customer'::tenant_type
         JOIN business_tenant_mappings btm2 ON btm2.id = ta1.source AND btm2.type = 'account'::tenant_type
         JOIN tenant_runtime_contexts ta2 ON ta1.id=ta2.id
         JOIN business_tenant_mappings btm3 ON btm3.id = ta2.tenant_id AND btm3.type = 'resource-group'::tenant_type
WHERE ta.tenant_id=ta1.tenant_id AND ta.source=ta1.source AND ta.id=ta1.id AND ta.owner = ta1.owner;

-- After fixing the swapped source and tenant_id there are leftover records stating that the GA knows about the resource from itself which is not correct as The GA knows about the resource from the CRM. Such records should be deleted
DELETE FROM tenant_runtime_contexts ta
    USING
        tenant_runtime_contexts ta2
            JOIN tenant_parents tp ON tp.parent_id = ta2.source AND tp.tenant_id=ta2.tenant_id
where ta.tenant_id =ta2.tenant_id AND ta.id=ta2.id AND ta.tenant_id = ta.source;

ALTER TABLE tenant_runtime_contexts
    ALTER COLUMN source SET NOT NULL;

ALTER TABLE tenant_runtime_contexts
    ADD PRIMARY KEY (tenant_id, id, source);

CREATE INDEX tenant_runtimes_contexts_source ON tenant_runtime_contexts (source);


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
       tp.tenant_id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id
         JOIN tenant_parents tp on ft.tenant_id = tp.parent_id;



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
       tp.tenant_id,
       true
FROM webhooks w
         JOIN formation_templates ft on w.formation_template_id = ft.id
         JOIN tenant_parents tp on ft.tenant_id = tp.parent_id;

-- Drop 'parent' column from 'business_tenant_mappings'
ALTER TABLE business_tenant_mappings
    DROP
        COLUMN parent;


DROP TRIGGER tenant_id_is_direct_parent_of_target_tenant_id ON automatic_scenario_assignments;
DROP FUNCTION IF EXISTS check_tenant_id_is_direct_parent_of_target_tenant_id();

CREATE
    OR REPLACE FUNCTION check_tenant_id_is_direct_parent_of_target_tenant_id() RETURNS TRIGGER AS
$$
DECLARE
    count INTEGER;
BEGIN
    EXECUTE format('SELECT COUNT(1) FROM tenant_parents WHERE tenant_id = %L AND parent_id = %L', NEW.target_tenant_id,
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


COMMIT;
