BEGIN;

UPDATE webhooks
SET input_template = '{"context":{"platform":"{{if .CustomerTenantContext.AccountID}}btp{{else}}unified-services{{end}}","uclFormationId":"{{.FormationID}}","accountId":"{{if .CustomerTenantContext.AccountID}}{{.CustomerTenantContext.AccountID}}{{else}}{{.CustomerTenantContext.Path}}{{end}}","crmId":"{{.CustomerTenantContext.CustomerID}}","operation":"{{.Operation}}"},"assignedTenant":{"state":"{{.Assignment.State}}","uclAssignmentId":"{{.Assignment.ID}}","deploymentRegion":"{{if .Application.Labels.region}}{{.Application.Labels.region}}{{else}}{{.ApplicationTemplate.Labels.region}}{{end}}","applicationNamespace":"{{if .Application.ApplicationNamespace}}{{.Application.ApplicationNamespace}}{{else}}{{.ApplicationTemplate.ApplicationNamespace}}{{end}}","applicationUrl":"{{.Application.BaseURL}}","applicationTenantId":"{{.Application.LocalTenantID}}","uclSystemName":"{{.Application.Name}}","uclSystemTenantId":"{{.Application.ID}}",{{if .ApplicationTemplate.Labels.parameters}}"parameters":{{.ApplicationTemplate.Labels.parameters}},{{end}}"configuration":{{.ReverseAssignment.Value}}},"receiverTenant":{"ownerTenants": [{{ Join .Runtime.Tenant.Parents }}],"state":"{{.ReverseAssignment.State}}","uclAssignmentId":"{{.ReverseAssignment.ID}}","deploymentRegion":"{{if and .RuntimeContext .RuntimeContext.Labels.region}}{{.RuntimeContext.Labels.region}}{{else}}{{.Runtime.Labels.region}}{{end}}","applicationNamespace":"{{.Runtime.ApplicationNamespace}}","applicationTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.Value}}{{else}}{{.Runtime.Labels.global_subaccount_id}}{{end}}","uclSystemTenantId":"{{if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}",{{if .Runtime.Labels.parameters}}"parameters":{{.Runtime.Labels.parameters}},{{end}}"configuration":{{.Assignment.Value}}}}'
WHERE runtime_id IN
      (SELECT runtime_id
       FROM labels
       WHERE runtime_id IS NOT NULL
         AND key = 'runtimeType'
         AND value = '"kyma"');

-- Copy business_tenant_mappings table
CREATE TABLE business_tenant_mappings_temp AS TABLE business_tenant_mappings;
CREATE INDEX business_tenant_mappings_temp_external_tenant ON business_tenant_mappings_temp (external_tenant);
CREATE INDEX business_tenant_mappings_temp_id ON business_tenant_mappings_temp (id);

-- Create many to many tenant_parents table
CREATE TABLE tenant_parents
(
    tenant_id uuid NOT NULL,
    parent_id uuid NOT NULL
);

-- Populate 'tenant_parents' table with data from 'business_tenant_mappings'
INSERT INTO tenant_parents (tenant_id, parent_id)
SELECT id, parent
FROM business_tenant_mappings_temp
WHERE parent IS NOT NULL;

-- Create key constraints
ALTER TABLE tenant_parents
    ADD CONSTRAINT tenant_parents_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_parents
    ADD CONSTRAINT tenant_parents_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_parents
    ADD PRIMARY KEY (tenant_id, parent_id);

-- Create indexes for tenant_parents table
CREATE INDEX tenant_parents_tenant_id ON tenant_parents (tenant_id);
CREATE INDEX tenant_parents_parent_id ON tenant_parents (parent_id);

-- Migrate tenant access records for applications
CREATE TABLE tenant_applications_temp AS TABLE tenant_applications;

-- Add source column to tenant_applications_temp table
ALTER TABLE tenant_applications_temp
    ADD COLUMN source uuid;

-- Add unique_tenant_applications_temp constraint
ALTER TABLE tenant_applications_temp
    ADD CONSTRAINT unique_tenant_applications_temp UNIQUE (tenant_id, id, source);

-- Add tenant access for the parents of the owning tenant
INSERT INTO tenant_applications_temp
    (SELECT tp.parent_id, ta.id, ta.owner, ta.tenant_id
     FROM tenant_applications_temp ta
              JOIN tenant_parents tp ON ta.tenant_id = tp.tenant_id)
ON CONFLICT (tenant_id,id,source) DO NOTHING;

-- Remove unique_tenant_applications_temp constraint
ALTER TABLE tenant_applications_temp
    DROP CONSTRAINT unique_tenant_applications_temp;

-- Create temporary table
CREATE TABLE temporary_tenant_applications_table AS TABLE tenant_applications_temp WITH NO DATA;

-- Populate the temporary table
INSERT INTO temporary_tenant_applications_table
SELECT parent_ta.*
FROM tenant_applications_temp parent_ta
         JOIN tenant_parents tp ON parent_ta.tenant_id = tp.parent_id
         JOIN tenant_applications_temp child_ta
              ON child_ta.tenant_id = tp.tenant_id AND parent_ta.ID = child_ta.ID;

-- Add tenant access records for the owning tenant itself
UPDATE tenant_applications_temp ta
SET source = ta.tenant_id
WHERE ta.source IS NULL
  AND NOT EXISTS (SELECT 1
                  FROM temporary_tenant_applications_table parent_ta
                  WHERE ta.tenant_id = parent_ta.tenant_id
                    AND ta.id = parent_ta.id);

-- Drop the temporary table
DROP TABLE temporary_tenant_applications_table;

-- Delete tenant access records with null source as they leftover access records for the parents of the owning tenant that were already populated by the previous queries
DELETE
FROM tenant_applications_temp
WHERE source IS NULL;

-- Tenant access records that originated from the directive were processed in the reverse direction. Now according to the TA table the CRM knows about the resource from the GA
-- We have to swap the source and tenant_id for records where the source is GA, the tenant_id is CRM and there is a record for the same resource where the tenant_id is resource-group (this means that the resource was created in Atom)

-- Create temporary table
CREATE TABLE temporary_tenant_applications_table AS TABLE tenant_applications_temp WITH NO DATA;

-- Populate the temporary table
INSERT INTO temporary_tenant_applications_table
SELECT ta1.*
FROM tenant_applications_temp ta1
         JOIN business_tenant_mappings_temp btm
              ON btm.id = ta1.tenant_id AND btm.type = 'customer'::tenant_type
         JOIN business_tenant_mappings_temp btm2 ON btm2.id = ta1.source AND btm2.type = 'account'::tenant_type
         JOIN tenant_applications_temp ta2 ON ta1.id = ta2.id
         JOIN business_tenant_mappings_temp btm3
              ON btm3.id = ta2.tenant_id AND btm3.type = 'resource-group'::tenant_type;


UPDATE tenant_applications_temp ta
SET source    = ta.tenant_id,
    tenant_id = ta.source
FROM temporary_tenant_applications_table tt
WHERE ta.tenant_id = tt.tenant_id
  AND ta.source = tt.source
  AND ta.id = tt.id
  AND ta.owner = tt.owner;

-- Drop the temporary table
DROP TABLE temporary_tenant_applications_table;

-- After fixing the swapped source and tenant_id there are leftover records stating that the GA knows about the resource from itself which is not correct as The GA knows about the resource from the CRM. Such records should be deleted
DELETE
FROM tenant_applications_temp ta USING
    tenant_applications_temp ta2
        JOIN tenant_parents tp
        ON tp.parent_id = ta2.source AND tp.tenant_id = ta2.tenant_id
WHERE ta.tenant_id = ta2.tenant_id
  AND ta.id = ta2.id
  AND ta.tenant_id = ta.source;

-- Migrate tenant access records for runtimes

-- Add source column to tenant_runtimes table
ALTER TABLE tenant_runtimes
    ADD COLUMN source uuid;

ALTER TABLE tenant_runtimes
    DROP CONSTRAINT tenant_runtimes_pkey;

ALTER TABLE tenant_runtimes
    ADD CONSTRAINT unique_tenant_runtimes UNIQUE (tenant_id, id, source);

-- Add tenant access for the parents of the owning tenant
INSERT INTO tenant_runtimes
    (SELECT tp.parent_id, ta.id, ta.owner, ta.tenant_id
     FROM tenant_runtimes ta
              JOIN tenant_parents tp ON ta.tenant_id = tp.tenant_id)
ON CONFLICT (tenant_id,id,source) DO NOTHING;

-- Add tenant access records for the owning tenant itself
UPDATE tenant_runtimes ta
SET source = ta.tenant_id
WHERE ta.source IS NULL
  AND NOT EXISTS (SELECT 1
                  FROM tenant_runtimes parent_ta
                           JOIN tenant_parents tp ON parent_ta.tenant_id = tp.parent_id
                           JOIN tenant_runtimes child_ta
                                ON child_ta.tenant_id = tp.tenant_id AND parent_ta.ID = child_ta.ID
                  WHERE ta.tenant_id = parent_ta.tenant_id
                    AND ta.id = parent_ta.id);

-- Delete tenant access records with null source as they leftover access records for the parents of the owning tenant that were already populated by the previous queries
DELETE
FROM tenant_runtimes
WHERE source IS NULL;

-- Tenant access records that originated from the directive were processed in the reverse direction. Now according to the TA table the CRM knows about the resource from the GA
-- We have to swap the source and tenant_id for records where the source is GA, the tenant_id is CRM and there is a record for the same resource where the tenant_id is resource-group (this means that the resource was created in Atom)

-- Create temporary table
CREATE TABLE temporary_tenant_runtimes_table AS TABLE tenant_runtimes WITH NO DATA;

-- Populate the temporary table
INSERT INTO temporary_tenant_runtimes_table
SELECT ta1.*
FROM tenant_runtimes ta1
         JOIN business_tenant_mappings_temp btm
              ON btm.id = ta1.tenant_id AND btm.type = 'customer'::tenant_type
         JOIN business_tenant_mappings_temp btm2 ON btm2.id = ta1.source AND btm2.type = 'account'::tenant_type
         JOIN tenant_runtimes ta2 ON ta1.id = ta2.id
         JOIN business_tenant_mappings_temp btm3
              ON btm3.id = ta2.tenant_id AND btm3.type = 'resource-group'::tenant_type;

UPDATE tenant_runtimes ta
SET source    = ta.tenant_id,
    tenant_id = ta.source
FROM temporary_tenant_runtimes_table tr
WHERE ta.tenant_id = tr.tenant_id
  AND ta.source = tr.source
  AND ta.id = tr.id
  AND ta.owner = tr.owner;

-- Drop the temporary table
DROP TABLE temporary_tenant_runtimes_table;

-- After fixing the swapped source and tenant_id there are leftover records stating that the GA knows about the resource from itself which is not correct as The GA knows about the resource from the CRM. Such records should be deleted
DELETE
FROM tenant_runtimes ta USING
    tenant_runtimes ta2
        JOIN tenant_parents tp
        ON tp.parent_id = ta2.source AND tp.tenant_id = ta2.tenant_id
WHERE ta.tenant_id = ta2.tenant_id
  AND ta.id = ta2.id
  AND ta.tenant_id = ta.source;

ALTER TABLE tenant_runtimes
    ADD CONSTRAINT tenant_runtimes_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;

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
    DROP CONSTRAINT tenant_runtime_contexts_pkey;

ALTER TABLE tenant_runtime_contexts
    ADD CONSTRAINT unique_tenant_runtime_contexts UNIQUE (tenant_id, id, source);

-- Add tenant access for the parents of the owning tenant
INSERT INTO tenant_runtime_contexts
    (SELECT tp.parent_id, ta.id, ta.owner, ta.tenant_id
     FROM tenant_runtime_contexts ta
              JOIN tenant_parents tp ON ta.tenant_id = tp.tenant_id)
ON CONFLICT (tenant_id,id,source) DO NOTHING;

-- Add tenant access records for the owning tenant itself
UPDATE tenant_runtime_contexts ta
SET source = ta.tenant_id
WHERE ta.source IS NULL
  AND NOT EXISTS (SELECT 1
                  FROM tenant_runtime_contexts parent_ta
                           JOIN tenant_parents tp ON parent_ta.tenant_id = tp.parent_id
                           JOIN tenant_runtime_contexts child_ta
                                ON child_ta.tenant_id = tp.tenant_id AND parent_ta.ID = child_ta.ID
                  WHERE ta.tenant_id = parent_ta.tenant_id
                    AND ta.id = parent_ta.id);

-- Delete tenant access records with null source as they leftover access records for the parents of the owning tenant that were already populated by the previous queries
DELETE
FROM tenant_runtime_contexts
WHERE source IS NULL;

-- Tenant access records that originated from the directive were processed in the reverse direction. Now according to the TA table the CRM knows about the resource from the GA
-- We have to swap the source and tenant_id for records where the source is GA, the tenant_id is CRM and there is a record for the same resource where the tenant_id is resource-group (this means that the resource was created in Atom)

-- Create temporary table
CREATE TABLE temporary_tenant_runtime_contexts_table AS TABLE tenant_runtime_contexts WITH NO DATA;

-- Populate the temporary table
INSERT INTO temporary_tenant_runtime_contexts_table
SELECT ta1.*
FROM tenant_runtime_contexts ta1
         JOIN business_tenant_mappings_temp btm
              ON btm.id = ta1.tenant_id AND btm.type = 'customer'::tenant_type
         JOIN business_tenant_mappings_temp btm2 ON btm2.id = ta1.source AND btm2.type = 'account'::tenant_type
         JOIN tenant_runtime_contexts ta2 ON ta1.id = ta2.id
         JOIN business_tenant_mappings_temp btm3
              ON btm3.id = ta2.tenant_id AND btm3.type = 'resource-group'::tenant_type;

UPDATE tenant_runtime_contexts ta
SET source    = ta.tenant_id,
    tenant_id = ta.source
FROM temporary_tenant_runtime_contexts_table trc
WHERE ta.tenant_id = trc.tenant_id
  AND ta.source = trc.source
  AND ta.id = trc.id
  AND ta.owner = trc.owner;

-- Drop the temporary table
DROP TABLE temporary_tenant_runtimes_table;

-- After fixing the swapped source and tenant_id there are leftover records stating that the GA knows about the resource from itself which is not correct as The GA knows about the resource from the CRM. Such records should be deleted
DELETE
FROM tenant_runtime_contexts ta USING
    tenant_runtime_contexts ta2
        JOIN tenant_parents tp
        ON tp.parent_id = ta2.source AND tp.tenant_id = ta2.tenant_id
WHERE ta.tenant_id = ta2.tenant_id
  AND ta.id = ta2.id
  AND ta.tenant_id = ta.source;

ALTER TABLE tenant_runtime_contexts
    ADD CONSTRAINT tenant_runtime_contexts_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;

ALTER TABLE tenant_runtime_contexts
    ALTER COLUMN source SET NOT NULL;

ALTER TABLE tenant_runtime_contexts
    ADD PRIMARY KEY (tenant_id, id, source);

CREATE INDEX tenant_runtimes_contexts_source ON tenant_runtime_contexts (source);

DROP TABLE business_tenant_mappings_temp;

-- Rename use the new tenant_applications table
LOCK business_tenant_mappings IN EXCLUSIVE MODE;

ALTER TABLE tenant_applications
    RENAME TO tenant_applications_old;
ALTER TABLE tenant_applications_old
    DROP CONSTRAINT tenant_applications_pkey;
ALTER TABLE tenant_applications_old
    DROP CONSTRAINT tenant_applications_id_fkey;
ALTER TABLE tenant_applications_old
    DROP CONSTRAINT tenant_applications_tenant_id_fkey;
ALTER
    INDEX tenant_applications_app_id RENAME TO tenant_applications_old_app_id;
ALTER
    INDEX tenant_applications_tenant_id RENAME TO tenant_applications_old_tenant_id;

ALTER TABLE tenant_applications_temp
    RENAME TO tenant_applications;
ALTER TABLE tenant_applications
    ADD PRIMARY KEY (tenant_id, id, source);

ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_tenant_id_fkey
        FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_id_fkey
        FOREIGN KEY (id) REFERENCES applications (id) ON DELETE CASCADE;
ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_source_fkey
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;

ALTER TABLE tenant_applications
    ALTER COLUMN source SET NOT NULL;

CREATE INDEX tenant_applications_app_id ON tenant_applications (id);
CREATE INDEX tenant_applications_tenant_id ON tenant_applications (tenant_id);
CREATE INDEX tenant_applications_source ON tenant_applications (source);

DROP VIEW IF EXISTS formation_templates_webhooks_tenants;

CREATE OR REPLACE VIEW formation_templates_webhooks_tenants
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
       tp.tenant_id,
       TRUE
FROM webhooks w
         JOIN formation_templates ft ON w.formation_template_id = ft.id
         JOIN tenant_parents tp ON ft.tenant_id = tp.parent_id;

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
       tp.tenant_id,
       TRUE
FROM webhooks w
         JOIN formation_templates ft ON w.formation_template_id = ft.id
         JOIN tenant_parents tp ON ft.tenant_id = tp.parent_id;

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
    EXECUTE FORMAT('SELECT COUNT(1) FROM tenant_parents WHERE tenant_id = %L AND parent_id = %L', NEW.target_tenant_id,
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
    CONSTRAINT TRIGGER tenant_id_is_direct_parent_of_target_tenant_id
    AFTER INSERT
    ON automatic_scenario_assignments
    FOR EACH ROW
EXECUTE PROCEDURE check_tenant_id_is_direct_parent_of_target_tenant_id();

DROP VIEW IF EXISTS tenants_packages;
DROP VIEW IF EXISTS tenants_products;
DROP VIEW IF EXISTS tenants_tombstones;
DROP VIEW IF EXISTS tombstones_tenants;
DROP VIEW IF EXISTS tenants_vendors;
DROP VIEW IF EXISTS vendors_tenants;
DROP VIEW IF EXISTS api_definitions_tenants;
DROP VIEW IF EXISTS integration_dependencies_tenants;
DROP VIEW IF EXISTS event_specifications_tenants;
DROP VIEW IF EXISTS event_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS event_api_definitions_tenants;
DROP VIEW IF EXISTS entity_types_tenants;
DROP VIEW IF EXISTS entity_type_mappings_tenants;
DROP VIEW IF EXISTS api_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS application_labels_tenants;
DROP VIEW IF EXISTS api_specifications_tenants;
DROP VIEW IF EXISTS bundle_instance_auths_tenants;
DROP VIEW IF EXISTS aspects_tenants;
DROP VIEW IF EXISTS aspect_event_resources_tenants;
DROP VIEW IF EXISTS application_webhooks_tenants;
DROP VIEW IF EXISTS capabilities_tenants;
DROP VIEW IF EXISTS bundles_tenants;
DROP VIEW IF EXISTS documents_tenants;
DROP VIEW IF EXISTS document_fetch_requests_tenants;
DROP VIEW IF EXISTS data_products_tenants;
DROP VIEW IF EXISTS capability_specifications_tenants;
DROP VIEW IF EXISTS capability_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS tenants_aspect_event_resources;
DROP VIEW IF EXISTS tenants_apps;
DROP VIEW IF EXISTS products_tenants;
DROP VIEW IF EXISTS packages_tenants;
DROP VIEW IF EXISTS labels_tenants;
DROP VIEW IF EXISTS tenants_integration_dependencies;
DROP VIEW IF EXISTS tenants_entity_types;
DROP VIEW IF EXISTS tenants_data_products;
DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_capabilities;
DROP VIEW IF EXISTS tenants_bundles;
DROP VIEW IF EXISTS tenants_aspects;
DROP VIEW IF EXISTS tenants_entity_type_mappings;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_apis;

CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, formation_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags,
             supported_use_cases, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible_supported, extensible_description,
             successors, resource_hash, documentation_labels, correlation_ids, direction, last_update, deprecation_date,
             responsible, usage)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                apis.id,
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
                apis.local_tenant_id,
                apis.short_description,
                apis.system_instance_aware,
                apis.policy_level,
                apis.custom_policy_level,
                apis.api_protocol,
                apis.tags,
                apis.supported_use_cases,
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
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                apis.successors,
                apis.resource_hash,
                apis.documentation_labels,
                apis.correlation_ids,
                apis.direction,
                apis.last_update,
                apis.deprecation_date,
                apis.responsible,
                apis.usage
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON apis.app_id = t_apps.id,
     JSONB_TO_RECORD(apis.extensible) actions(supported text, description text);

CREATE OR REPLACE VIEW tenants_events
            (tenant_id, formation_id, id, app_id, name, description, group_name, version_value, version_deprecated,
             version_deprecated_since, version_for_removal, ord_id, local_tenant_id, short_description,
             system_instance_aware, policy_level, custom_policy_level, changelog_entries, links, tags, countries,
             release_status, sunset_date, labels, package_id, visibility, disabled, part_of_products, line_of_business,
             industry, ready, created_at, updated_at, deleted_at, error, implementation_standard,
             custom_implementation_standard, custom_implementation_standard_description, extensible_supported,
             extensible_description, successors, resource_hash, correlation_ids, last_update, deprecation_date,
             event_resource_links, responsible)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                events.id,
                events.app_id,
                events.name,
                events.description,
                events.group_name,
                events.version_value,
                events.version_deprecated,
                events.version_deprecated_since,
                events.version_for_removal,
                events.ord_id,
                events.local_tenant_id,
                events.short_description,
                events.system_instance_aware,
                events.policy_level,
                events.custom_policy_level,
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
                events.implementation_standard,
                events.custom_implementation_standard,
                events.custom_implementation_standard_description,
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                events.successors,
                events.resource_hash,
                events.correlation_ids,
                events.last_update,
                events.deprecation_date,
                events.event_resource_links,
                events.responsible
FROM event_api_definitions events
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON events.app_id = t_apps.id,
     JSONB_TO_RECORD(events.extensible) actions(supported text, description text);

CREATE OR REPLACE VIEW tenants_entity_type_mappings
            (tenant_id, id, api_definition_id, event_definition_id, api_model_SELECTors, entity_type_targets)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                etm.id,
                etm.api_definition_id,
                etm.event_definition_id,
                etm.api_model_SELECTors,
                etm.entity_type_targets
FROM entity_type_mappings etm
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e) t_api_event_def
              ON etm.api_definition_id = t_api_event_def.id OR etm.event_definition_id = t_api_event_def.id;

CREATE OR REPLACE VIEW api_definitions_tenants AS
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
       apis.local_tenant_id,
       apis.short_description,
       apis.system_instance_aware,
       apis.policy_level,
       apis.custom_policy_level,
       apis.api_protocol,
       apis.tags,
       apis.supported_use_cases,
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
       apis.correlation_ids,
       apis.direction,
       apis.last_update,
       apis.deprecation_date,
       apis.responsible,
       apis.usage,
       ta.tenant_id,
       ta.owner
FROM api_definitions AS apis
         INNER JOIN tenant_applications ta ON ta.id = apis.app_id;

CREATE OR REPLACE VIEW api_specifications_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner
FROM fetch_requests AS fr
         INNER JOIN specifications s ON fr.spec_id = s.id
         INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
         INNER JOIN tenant_applications ta ON ta.id = ad.app_id;

CREATE OR REPLACE VIEW api_specifications_tenants AS
(
SELECT s.*, ta.tenant_id, ta.owner
FROM specifications AS s
         INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
         INNER JOIN tenant_applications ta ON ta.id = ad.app_id);

CREATE OR REPLACE VIEW application_labels_tenants AS
SELECT l.id, ta.tenant_id, ta.owner
FROM labels AS l
         INNER JOIN tenant_applications ta
                    ON l.app_id = ta.id AND (l.tenant_id IS NULL OR l.tenant_id = ta.tenant_id);

CREATE OR REPLACE VIEW application_webhooks_tenants
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
         JOIN tenant_applications ta ON w.app_id = ta.id;

CREATE OR REPLACE VIEW aspect_event_resources_tenants AS
SELECT a.*, ta.tenant_id, ta.owner
FROM aspect_event_resources AS a
         INNER JOIN tenant_applications ta ON ta.id = a.app_id;

CREATE OR REPLACE VIEW aspects_tenants AS
SELECT a.*, ta.tenant_id, ta.owner
FROM aspects AS a
         INNER JOIN tenant_applications ta ON ta.id = a.app_id;

CREATE OR REPLACE VIEW bundle_instance_auths_tenants AS
SELECT bia.*, ta.tenant_id, ta.owner
FROM bundle_instance_auths AS bia
         INNER JOIN bundles b ON b.id = bia.bundle_id
         INNER JOIN tenant_applications ta ON ta.id = b.app_id;

CREATE OR REPLACE VIEW bundles_tenants AS
SELECT b.id,
       b.app_id,
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
       b.tags,
       b.version,
       b.resource_hash,
       b.local_tenant_id,
       b.app_template_version_id,
       b.last_update,
       ta.tenant_id,
       ta.owner
FROM bundles AS b
         INNER JOIN tenant_applications ta ON ta.id = b.app_id;

CREATE OR REPLACE VIEW capabilities_tenants AS
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
       c.related_entity_types,
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

CREATE OR REPLACE VIEW capability_specifications_fetch_requests_tenants AS
(
SELECT fr.*, ta.tenant_id, ta.owner
FROM fetch_requests AS fr
         INNER JOIN specifications s ON fr.spec_id = s.id
         INNER JOIN capabilities AS cd ON cd.id = s.capability_def_id
         INNER JOIN tenant_applications ta ON ta.id = cd.app_id);

CREATE OR REPLACE VIEW capability_specifications_tenants AS
(
SELECT s.*, ta.tenant_id, ta.owner
FROM specifications AS s
         INNER JOIN capabilities AS cd ON cd.id = s.capability_def_id
         INNER JOIN tenant_applications ta ON ta.id = cd.app_id);

CREATE OR REPLACE VIEW data_products_tenants AS
SELECT d.*, ta.tenant_id, ta.owner
FROM data_products AS d
         INNER JOIN tenant_applications ta ON ta.id = d.app_id;

CREATE OR REPLACE VIEW document_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner
FROM fetch_requests AS fr
         INNER JOIN documents d ON fr.document_id = d.id
         INNER JOIN tenant_applications ta ON ta.id = d.app_id;

CREATE OR REPLACE VIEW documents_tenants AS
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
       d.app_template_version_id,
       ta.tenant_id,
       ta.owner
FROM documents AS d
         INNER JOIN tenant_applications ta ON ta.id = d.app_id;


CREATE OR REPLACE VIEW entity_type_mappings_tenants(id, tenant_id, owner)
AS
SELECT DISTINCT etm.id,
                t_api_event_def.tenant_id,
                t_api_event_def.owner
FROM entity_type_mappings etm
         JOIN (SELECT a.id,
                      a.tenant_id,
                      ta.owner
               FROM tenants_apis a
                        JOIN tenant_applications ta ON ta.id = a.app_id
               UNION ALL
               SELECT e.id,
                      e.tenant_id,
                      ta.owner
               FROM tenants_events e
                        JOIN tenant_applications ta ON ta.id = e.app_id) t_api_event_def
              ON etm.api_definition_id = t_api_event_def.id OR etm.event_definition_id = t_api_event_def.id;

CREATE OR REPLACE VIEW entity_types_tenants AS
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
       et.deprecation_date,
       ta.tenant_id,
       ta.owner
FROM entity_types AS et
         INNER JOIN tenant_applications ta ON ta.id = et.app_id;

CREATE OR REPLACE VIEW event_api_definitions_tenants AS
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
       events.local_tenant_id,
       events.short_description,
       events.system_instance_aware,
       events.policy_level,
       events.custom_policy_level,
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
       events.implementation_standard,
       events.custom_implementation_standard,
       events.custom_implementation_standard_description,
       events.successors,
       events.resource_hash,
       events.correlation_ids,
       events.last_update,
       events.deprecation_date,
       events.event_resource_links,
       events.responsible,
       ta.tenant_id,
       ta.owner
FROM event_api_definitions AS events
         INNER JOIN tenant_applications ta ON ta.id = events.app_id;

CREATE OR REPLACE VIEW event_specifications_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner
FROM fetch_requests AS fr
         INNER JOIN specifications s ON fr.spec_id = s.id
         INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
         INNER JOIN tenant_applications ta ON ta.id = ead.app_id;

CREATE OR REPLACE VIEW event_specifications_tenants AS
(
SELECT s.*, ta.tenant_id, ta.owner
FROM specifications AS s
         INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
         INNER JOIN tenant_applications ta ON ta.id = ead.app_id);

CREATE OR REPLACE VIEW integration_dependencies_tenants AS
SELECT i.*, ta.tenant_id, ta.owner
FROM integration_dependencies AS i
         INNER JOIN tenant_applications ta ON ta.id = i.app_id;

CREATE OR REPLACE VIEW labels_tenants AS
(SELECT l.id, ta.tenant_id, ta.owner
 FROM labels AS l
          INNER JOIN tenant_applications ta
                     ON l.app_id = ta.id AND (l.tenant_id IS NULL OR l.tenant_id = ta.tenant_id))
UNION ALL
(SELECT l.id, tr.tenant_id, tr.owner
 FROM labels AS l
          INNER JOIN tenant_runtimes tr
                     ON l.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id))
UNION ALL
(SELECT l.id, trc.tenant_id, trc.owner
 FROM labels AS l
          INNER JOIN tenant_runtime_contexts trc
                     ON l.runtime_context_id = trc.id AND (l.tenant_id IS NULL OR l.tenant_id = trc.tenant_id));

CREATE OR REPLACE VIEW packages_tenants
            (id, ord_id, title, short_description, description, version, package_links, links, licence_type, tags,
             countries, labels, policy_level, app_id, custom_policy_level, vendor, part_of_products, line_of_business,
             industry, resource_hash, documentation_labels, support_info, tenant_id, owner)
AS
SELECT p.id,
       p.ord_id,
       p.title,
       p.short_description,
       p.description,
       p.version,
       p.package_links,
       p.links,
       p.licence_type,
       p.tags,
       p.countries,
       p.labels,
       p.policy_level,
       p.app_id,
       p.custom_policy_level,
       p.vendor,
       p.part_of_products,
       p.line_of_business,
       p.industry,
       p.resource_hash,
       p.documentation_labels,
       p.support_info,
       ta.tenant_id,
       ta.owner
FROM packages p
         JOIN tenant_applications ta ON ta.id = p.app_id;

CREATE OR REPLACE VIEW products_tenants AS
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
       p.tags,
       p.app_template_version_id,
       p.description,
       ta.tenant_id,
       ta.owner
FROM products AS p
         INNER JOIN tenant_applications AS ta ON ta.id = p.app_id;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, formation_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, tags, ready, created_at, updated_at, deleted_at,
             error, app_template_id, correlation_ids, system_number, application_namespace, local_tenant_id,
             tenant_business_type_id, product_type)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                apps.id,
                apps.name,
                apps.description,
                apps.status_condition,
                apps.status_timestamp,
                apps.healthcheck_url,
                apps.integration_system_id,
                apps.provider_name,
                apps.base_url,
                apps.labels,
                apps.tags,
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                COALESCE(apps.application_namespace, tmpl.application_namespace) AS application_namespace,
                apps.local_tenant_id,
                apps.tenant_business_type_id,
                tmpl.name                                                        AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON apps.id = t_apps.id;

CREATE OR REPLACE VIEW tenants_aspect_event_resources
            (tenant_id, formation_id, id, aspect_id, app_id, ord_id, min_version, subset, ready, created_at, updated_at,
             deleted_at, error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                a.id,
                a.aspect_id,
                a.app_id,
                a.ord_id,
                a.min_version,
                a.subset,
                a.ready,
                a.created_at,
                a.updated_at,
                a.deleted_at,
                a.error
FROM aspect_event_resources a
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON a.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_aspects
            (tenant_id, formation_id, id, integration_dependency_id, app_id, title, description, mandatory,
             support_multiple_providers, api_resources, ready, created_at, updated_at, deleted_at,
             error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                a.id,
                a.integration_dependency_id,
                a.app_id,
                a.title,
                a.description,
                a.mandatory,
                a.support_multiple_providers,
                a.api_resources,
                a.ready,
                a.created_at,
                a.updated_at,
                a.deleted_at,
                a.error
FROM aspects a
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON a.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_bundles
            (tenant_id, formation_id, id, app_id, name, description, version, instance_auth_request_json_schema,
             default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, tags,
             credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, last_update,
             resource_hash)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                b.id,
                b.app_id,
                b.name,
                b.description,
                b.version,
                b.instance_auth_request_json_schema,
                b.default_instance_auth,
                b.ord_id,
                b.local_tenant_id,
                b.short_description,
                b.links,
                b.labels,
                b.tags,
                b.credential_exchange_strategies,
                b.ready,
                b.created_at,
                b.updated_at,
                b.deleted_at,
                b.error,
                b.correlation_ids,
                b.last_update,
                b.resource_hash
FROM bundles b
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON b.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_capabilities
            (tenant_id, formation_id, id, app_id, name, description, type, custom_type, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ord_id, local_tenant_id,
             short_description, system_instance_aware, tags, related_entity_types, links, release_status, labels,
             package_id, visibility, ready, created_at, updated_at, deleted_at, error, resource_hash,
             documentation_labels, correlation_ids, last_update)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                c.id,
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
                c.related_entity_types,
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
                c.last_update
FROM capabilities c
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON c.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format,
             event_spec_type, capability_def_id, capability_spec_type, capability_spec_format, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_capability_def.tenant_id,
                spec.id,
                spec.api_def_id,
                spec.event_def_id,
                spec.spec_data,
                spec.api_spec_format,
                spec.api_spec_type,
                spec.event_spec_format,
                spec.event_spec_type,
                spec.capability_def_id,
                spec.capability_spec_type,
                spec.capability_spec_format,
                spec.custom_type,
                spec.created_at
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e
               UNION ALL
               SELECT c.id,
                      c.tenant_id
               FROM tenants_capabilities c) t_api_event_capability_def
              ON spec.api_def_id = t_api_event_capability_def.id OR spec.event_def_id = t_api_event_capability_def.id OR
                 spec.capability_def_id = t_api_event_capability_def.id;

CREATE OR REPLACE VIEW tenants_data_products
            (tenant_id, formation_id, id, app_id, ord_id, local_tenant_id, correlation_ids, title, short_description,
             description, package_id, last_update,
             visibility, release_status, disabled, deprecation_date, sunset_date, successors, changelog_entries, type,
             category, entity_types, input_ports,
             output_ports, responsible, data_product_links, links, industry, line_of_business, tags, labels,
             documentation_labels, policy_level, custom_policy_level,
             system_instance_aware, resource_hash, version_value, version_deprecated, version_deprecated_since,
             version_for_removal, ready, created_at, updated_at, deleted_at, error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                d.id,
                d.app_id,
                d.ord_id,
                d.local_tenant_id,
                d.correlation_ids,
                d.title,
                d.short_description,
                d.description,
                d.package_id,
                d.last_update,
                d.visibility,
                d.release_status,
                d.disabled,
                d.deprecation_date,
                d.sunset_date,
                d.successors,
                d.changelog_entries,
                d.type,
                d.category,
                d.entity_types,
                d.input_ports,
                d.output_ports,
                d.responsible,
                d.data_product_links,
                d.links,
                d.industry,
                d.line_of_business,
                d.tags,
                d.labels,
                d.documentation_labels,
                d.policy_level,
                d.custom_policy_level,
                d.system_instance_aware,
                d.resource_hash,
                d.version_value,
                d.version_deprecated,
                d.version_deprecated_since,
                d.version_for_removal,
                d.ready,
                d.created_at,
                d.updated_at,
                d.deleted_at,
                d.error
FROM data_products d
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps ON d.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_entity_types
            (tenant_id, formation_id, id, ord_id, app_id, local_tenant_id, level, title, short_description, description,
             system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update,
             policy_level, custom_policy_level, release_status, sunset_date, successors, extensible_supported,
             extensible_description, tags, labels, documentation_labels, resource_hash, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, deprecation_date)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                et.id,
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
                actions.supported   AS extensible_supported,
                actions.description AS extensible_description,
                et.tags,
                et.labels,
                et.documentation_labels,
                et.resource_hash,
                et.version_value,
                et.version_deprecated,
                et.version_deprecated_since,
                et.version_for_removal,
                et.deprecation_date
FROM entity_types et
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON et.app_id = t_apps.id,
     JSONB_TO_RECORD(et.extensible) actions(supported text, description text);

CREATE OR REPLACE VIEW tenants_integration_dependencies
            (tenant_id, formation_id, id, app_id, ord_id, local_tenant_id, correlation_ids, title, short_description,
             description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory,
             related_integration_dependencies, links, tags, labels, documentation_labels, resource_hash, version_value,
             version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at,
             deleted_at, error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                i.id,
                i.app_id,
                i.ord_id,
                i.local_tenant_id,
                i.correlation_ids,
                i.title,
                i.short_description,
                i.description,
                i.package_id,
                i.last_update,
                i.visibility,
                i.release_status,
                i.sunset_date,
                i.successors,
                i.mandatory,
                i.related_integration_dependencies,
                i.links,
                i.tags,
                i.labels,
                i.documentation_labels,
                i.resource_hash,
                i.version_value,
                i.version_deprecated,
                i.version_deprecated_since,
                i.version_for_removal,
                i.ready,
                i.created_at,
                i.updated_at,
                i.deleted_at,
                i.error
FROM integration_dependencies i
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON i.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, formation_id, id, ord_id, title, short_description, description, version, package_links, links,
             licence_type, tags, runtime_restriction, countries, labels, policy_level, app_id, custom_policy_level,
             vendor, part_of_products,
             line_of_business, industry, resource_hash, support_info)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                p.id,
                p.ord_id,
                p.title,
                p.short_description,
                p.description,
                p.version,
                p.package_links,
                p.links,
                p.licence_type,
                p.tags,
                p.runtime_restriction,
                p.countries,
                p.labels,
                p.policy_level,
                p.app_id,
                p.custom_policy_level,
                p.vendor,
                p.part_of_products,
                p.line_of_business,
                p.industry,
                p.resource_hash,
                p.support_info
FROM packages p
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON p.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_products
            (tenant_id, formation_id, ord_id, app_id, title, short_description, vendor, parent, labels, tags,
             correlation_ids, id, documentation_labels, description)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                p.ord_id,
                p.app_id,
                p.title,
                p.short_description,
                p.vendor,
                p.parent,
                p.labels,
                p.tags,
                p.correlation_ids,
                p.id,
                p.documentation_labels,
                p.description
FROM products p
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON p.app_id = t_apps.id OR p.app_id IS NULL;

CREATE OR REPLACE VIEW tenants_tombstones (tenant_id, formation_id, ord_id, app_id, removal_date, id, description)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                t.ord_id,
                t.app_id,
                t.removal_date,
                t.id,
                t.description
FROM tombstones t
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON t.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_vendors
            (tenant_id, formation_id, ord_id, app_id, title, labels, tags, partners, id, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                v.ord_id,
                v.app_id,
                v.title,
                v.labels,
                v.tags,
                v.partners,
                v.id,
                v.documentation_labels
FROM vendors v
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON v.app_id = t_apps.id OR v.app_id IS NULL;

CREATE OR REPLACE VIEW tombstones_tenants AS
SELECT t.ord_id,
       t.app_id,
       t.removal_date,
       t.id,
       t.app_template_version_id,
       t.description,
       ta.tenant_id,
       ta.owner
FROM tombstones AS t
         INNER JOIN tenant_applications AS ta ON ta.id = t.app_id;

CREATE OR REPLACE VIEW vendors_tenants AS
SELECT v.ord_id,
       v.app_id,
       v.title,
       v.labels,
       v.partners,
       v.id,
       v.documentation_labels,
       v.tags,
       v.app_template_version_id,
       ta.tenant_id,
       ta.owner
FROM vendors AS v
         INNER JOIN tenant_applications AS ta ON ta.id = v.app_id;

COMMIT;
