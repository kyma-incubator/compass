BEGIN;

UPDATE webhooks
SET input_template = '{"context":{"platform":"{{if .CustomerTenantContext.AccountID}}btp{{else}}unified-services{{end}}","uclFormationId":"{{.FormationID}}","accountId":"{{if .CustomerTenantContext.AccountID}}{{.CustomerTenantContext.AccountID}}{{else}}{{.CustomerTenantContext.Path}}{{end}}","crmId":"{{.CustomerTenantContext.CustomerID}}","operation":"{{.Operation}}"},"assignedTenant":{"state":"{{.Assignment.State}}","uclAssignmentId":"{{.Assignment.ID}}","deploymentRegion":"{{if .Application.Labels.region}}{{.Application.Labels.region}}{{else}}{{.ApplicationTemplate.Labels.region}}{{end}}","applicationNamespace":"{{if .Application.ApplicationNamespace}}{{.Application.ApplicationNamespace}}{{else}}{{.ApplicationTemplate.ApplicationNamespace}}{{end}}","applicationUrl":"{{.Application.BaseURL}}","applicationTenantId":"{{.Application.LocalTenantID}}","uclSystemName":"{{.Application.Name}}","uclSystemTenantId":"{{.Application.ID}}",{{if .ApplicationTemplate.Labels.parameters}}"parameters":{{.ApplicationTemplate.Labels.parameters}},{{end}}"configuration":{{.ReverseAssignment.Value}}},"receiverTenant":{"ownerTenant":"{{.Runtime.Tenant.Parent}}","state":"{{.ReverseAssignment.State}}","uclAssignmentId":"{{.ReverseAssignment.ID}}","deploymentRegion":"{{if and .RuntimeContext .RuntimeContext.Labels.region}}{{.RuntimeContext.Labels.region}}{{else}}{{.Runtime.Labels.region}}{{end}}","applicationNamespace":"{{.Runtime.ApplicationNamespace}}","applicationTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.Value}}{{else}}{{.Runtime.Labels.global_subaccount_id}}{{end}}","uclSystemTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}",{{if .Runtime.Labels.parameters}}"parameters":{{.Runtime.Labels.parameters}},{{end}}"configuration":{{.Assignment.Value}}}}'
WHERE runtime_id IN
      (SELECT runtime_id
       FROM labels
       WHERE runtime_id IS NOT NULL
         AND
    KEY = 'runtimeType'
  AND VALUE = '"kyma"'
    );

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
WHERE t1.tenant_id = t2.tenant_id AND t1.id=t2.id AND t1.owner= FALSE AND t2.owner= TRUE;

ALTER TABLE tenant_applications DROP COLUMN SOURCE;
ALTER TABLE tenant_applications
    ADD PRIMARY KEY (tenant_id, id);

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
WHERE t1.tenant_id = t2.tenant_id AND t1.id=t2.id AND t1.owner= FALSE AND t2.owner= TRUE;

ALTER TABLE tenant_runtimes DROP COLUMN SOURCE;

ALTER TABLE tenant_runtimes
    ADD PRIMARY KEY (tenant_id, id);

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
WHERE t1.tenant_id = t2.tenant_id AND t1.id=t2.id AND t1.owner= FALSE AND t2.owner= TRUE;

ALTER TABLE tenant_runtime_contexts DROP COLUMN SOURCE;
ALTER TABLE tenant_runtime_contexts
    ADD PRIMARY KEY (tenant_id, id);

-- Add parent column to business tenant mapping
ALTER TABLE business_tenant_mappings
    ADD COLUMN parent uuid;

DROP VIEW IF EXISTS formation_templates_webhooks_tenants;

CREATE
OR REPLACE VIEW formation_templates_webhooks_tenants (id, app_id, url, TYPE, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
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
       TRUE
FROM webhooks w
         JOIN formation_templates ft ON w.formation_template_id = ft.id
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
       TRUE
FROM webhooks w
         JOIN formation_templates ft ON w.formation_template_id = ft.id
         JOIN business_tenant_mappings btm ON ft.tenant_id = btm.parent;

DROP VIEW IF EXISTS webhooks_tenants;
CREATE
OR REPLACE VIEW webhooks_tenants
            (id, app_id, url, TYPE, auth, mode, correlation_id_key, retry_interval, timeout, url_template,
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
       TRUE
FROM webhooks w
         JOIN formation_templates ft ON w.formation_template_id = ft.id
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
       TRUE
FROM webhooks w
         JOIN formation_templates ft ON w.formation_template_id = ft.id
         JOIN business_tenant_mappings btm ON ft.tenant_id = btm.parent;

-- Drop indexes for tenant_parents table
DROP INDEX IF EXISTS tenant_parents_tenant_id;
DROP INDEX IF EXISTS tenant_parents_parent_id;

-- Drop tenant_parents table
DROP TABLE IF EXISTS tenant_parents;

DROP VIEW IF EXISTS api_definitions_tenants;
CREATE
OR REPLACE VIEW api_definitions_tenants AS
SELECT apis.id,
       apis.app_id,
       apis.name,
       apis.description,
       apis.group_name,
       apis.default_auth,
       apis.version_value,
       apis.version_deprecated,
       apis.version_deprecated_since,
       apis.version_for_removal,
       apis.ord_id,
       apis.short_description,
       apis.system_instance_aware,
       apis.api_protocol,
       apis.tags,
       apis.countries,
       apis.links,
       apis.api_resource_links,
       apis.release_status,
       apis.sunset_date,
       apis.changelog_entries,
       apis.labels,
       apis.package_id,
       apis.visibility,
       apis.disabled,
       apis.part_of_products,
       apis.line_of_business,
       apis.industry,
       apis.ready,
       apis.created_at,
       apis.updated_at,
       apis.deleted_at,
       apis.error,
       apis.implementation_standard,
       apis.custom_implementation_standard,
       apis.custom_implementation_standard_description,
       apis.target_urls,
       apis.successors,
       apis.resource_hash,
       apis.documentation_labels,
       ta.tenant_id,
       ta.owner
FROM api_definitions AS apis
         INNER JOIN tenant_applications ta ON ta.id = apis.app_id;


DROP VIEW IF EXISTS event_api_definitions_tenants;
CREATE
OR REPLACE VIEW event_api_definitions_tenants AS
SELECT events.id,
       events.app_id,
       events.name,
       events.description,
       events.group_name,
       events.version_value,
       events.version_deprecated,
       events.version_deprecated_since,
       events.version_for_removal,
       events.ord_id,
       events.short_description,
       events.system_instance_aware,
       events.changelog_entries,
       events.links,
       events.tags,
       events.countries,
       events.release_status,
       events.sunset_date,
       events.labels,
       events.package_id,
       events.visibility,
       events.disabled,
       events.part_of_products,
       events.line_of_business,
       events.industry,
       events.ready,
       events.created_at,
       events.updated_at,
       events.deleted_at,
       events.error,
       events.successors,
       events.resource_hash,
       ta.tenant_id,
       ta.owner
FROM event_api_definitions AS events
         INNER JOIN tenant_applications ta ON ta.id = events.app_id;

DROP VIEW IF EXISTS capabilities_tenants;
CREATE
OR REPLACE VIEW capabilities_tenants AS
SELECT c.id,
       c.app_id,
       c.name,
       c.description,
       c.type,
       c.custom_type,
       c.version_value,
       c.version_deprecated,
       c.version_deprecated_since,
       c.version_for_removal,
       c.ord_id,
       c.local_tenant_id,
       c.short_description,
       c.system_instance_aware,
       c.tags,
       c.links,
       c.release_status,
       c.labels,
       c.package_id,
       c.visibility,
       c.ready,
       c.created_at,
       c.updated_at,
       c.deleted_at,
       c.error,
       c.resource_hash,
       c.documentation_labels,
       c.correlation_ids,
       c.last_update,
       ta.tenant_id,
       ta.owner
FROM capabilities AS c
         INNER JOIN tenant_applications ta ON ta.id = c.app_id;

DROP VIEW IF EXISTS entity_types_tenants;
CREATE
OR REPLACE VIEW entity_types_tenants AS
SELECT et.id,
       et.ord_id,
       et.app_id,
       et.local_tenant_id,
       et.level,
       et.title,
       et.short_description,
       et.description,
       et.system_instance_aware,
       et.changelog_entries,
       et.package_id,
       et.visibility,
       et.links,
       et.part_of_products,
       et.last_update,
       et.policy_level,
       et.custom_policy_level,
       et.release_status,
       et.sunset_date,
       et.successors,
       et.tags,
       et.labels,
       et.documentation_labels,
       et.resource_hash,
       et.version_value,
       et.version_deprecated,
       et.version_deprecated_since,
       et.version_for_removal,
       ta.tenant_id,
       ta.owner
FROM entity_types AS et
         INNER JOIN tenant_applications ta ON ta.id = et.app_id;

DROP VIEW IF EXISTS products_tenants;
CREATE
OR REPLACE VIEW products_tenants AS
SELECT p.ord_id,
       p.app_id,
       p.title,
       p.short_description,
       p.vendor,
       p.parent,
       p.labels,
       p.correlation_ids,
       p.id,
       p.documentation_labels,
       ta.tenant_id,
       ta.owner
FROM products AS p
         INNER JOIN tenant_applications AS ta ON ta.id = p.app_id;

DROP VIEW IF EXISTS tombstones_tenants;
CREATE
OR REPLACE VIEW tombstones_tenants AS
SELECT t.ord_id,
       t.app_id,
       t.removal_date,
       t.id,
       ta.tenant_id,
       ta.owner
FROM tombstones AS t
         INNER JOIN tenant_applications AS ta ON ta.id = t.app_id;

DROP VIEW IF EXISTS bundles_tenants;
CREATE
OR REPLACE VIEW bundles_tenants AS
SELECT b.app_id,
       b.name,
       b.description,
       b.instance_auth_request_json_schema,
       b.default_instance_auth,
       b.ord_id,
       b.short_description,
       b.links,
       b.labels,
       b.credential_exchange_strategies,
       b.ready,
       b.created_at,
       b.updated_at,
       b.deleted_at,
       b.error,
       b.correlation_ids,
       b.documentation_labels,
       ta.tenant_id,
       ta.owner
FROM bundles AS b
         INNER JOIN tenant_applications ta ON ta.id = b.app_id;

DROP VIEW IF EXISTS documents_tenants;
CREATE
OR REPLACE VIEW documents_tenants AS
SELECT d.id,
       d.app_id,
       d.title,
       d.display_name,
       d.description,
       d.format,
       d.kind,
       d.data,
       d.bundle_id,
       d.ready,
       d.created_at,
       d.updated_at,
       d.deleted_at,
       d.error,
       ta.tenant_id,
       ta.owner
FROM documents AS d
         INNER JOIN tenant_applications ta ON ta.id = d.app_id;

DROP VIEW IF EXISTS vendors_tenants;
CREATE
OR REPLACE VIEW vendors_tenants AS
SELECT v.ord_id,
       v.app_id,
       v.title,
       v.labels,
       v.partners,
       v.id,
       ta.tenant_id,
       ta.owner
FROM vendors AS v
         INNER JOIN tenant_applications AS ta ON ta.id = v.app_id;

LOCK
business_tenant_mappings IN EXCLUSIVE MODE;

-- Fill parent column
UPDATE business_tenant_mappings
SET parent=parent_id FROM tenant_parents
WHERE business_tenant_mappings.id = tenant_parents.tenant_id
  AND business_tenant_mappings.type <> 'cost-object'::tenant_type;

-- Add business tenant mapping parent fk
ALTER TABLE business_tenant_mappings
    ADD CONSTRAINT business_tenant_mappings_parent_fk
        FOREIGN KEY (parent)
            REFERENCES business_tenant_mappings (id);

-- Create parent index
CREATE INDEX parent_index ON business_tenant_mappings (parent);

-- Create tenant_id_is_direct_parent_of_target_tenant_id trigger
DROP TRIGGER tenant_id_is_direct_parent_of_target_tenant_id ON automatic_scenario_assignments;
DROP FUNCTION IF EXISTS check_tenant_id_is_direct_parent_of_target_tenant_id();

CREATE
OR REPLACE FUNCTION check_tenant_id_is_direct_parent_of_target_tenant_id() RETURNS TRIGGER AS
$$
DECLARE
COUNT INTEGER;
BEGIN
EXECUTE format('SELECT COUNT(1) FROM business_tenant_mappings WHERE id = %L AND parent = %L', NEW.target_tenant_id,
               NEW.tenant_id) INTO COUNT;
IF
COUNT = 0 THEN
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
