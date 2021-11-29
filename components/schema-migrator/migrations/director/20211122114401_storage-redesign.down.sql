BEGIN;

-- This is not exhaustive down migration as the storage redesign UP migration involves a data loss.
-- This is just DDL down migration. Consider restoring from backup instead of using DOWN migration.
-- If down migration is executed the lost information will be:
--     - All ASAs for key != global_subaccount_id OR value not valid subaccount in our DB
--     - All label definitions with key != scenario.

ALTER TABLE label_definitions DROP CONSTRAINT key_is_scenario;
DROP TRIGGER tenant_id_is_direct_parent_of_target_tenant_id ON automatic_scenario_assignments;
DROP FUNCTION IF EXISTS check_tenant_id_is_direct_parent_of_target_tenant_id();

ALTER TABLE automatic_scenario_assignments ADD COLUMN selector_key VARCHAR(256);
ALTER TABLE automatic_scenario_assignments ADD COLUMN selector_value VARCHAR(256);

UPDATE automatic_scenario_assignments asa
SET selector_key = 'global_subaccount_id',
    selector_value = (SELECT external_tenant FROM business_tenant_mappings WHERE id = asa.target_tenant_id);

ALTER TABLE automatic_scenario_assignments DROP COLUMN target_tenant_id;

DROP VIEW IF EXISTS runtime_webhooks_tenants;
DROP VIEW IF EXISTS application_webhooks_tenants;
DROP VIEW IF EXISTS webhooks_tenants;
DROP VIEW IF EXISTS vendors_tenants;
DROP VIEW IF EXISTS tombstones_tenants;
DROP VIEW IF EXISTS products_tenants;
DROP VIEW IF EXISTS packages_tenants;
DROP VIEW IF EXISTS runtime_contexts_tenants;
DROP VIEW IF EXISTS runtime_labels_tenants;
DROP VIEW IF EXISTS application_labels_tenants;
DROP VIEW IF EXISTS labels_tenants;
DROP VIEW IF EXISTS runtime_contexts_labels_tenants;
DROP VIEW IF EXISTS event_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS api_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS document_fetch_requests_tenants;
DROP VIEW IF EXISTS event_specifications_tenants;
DROP VIEW IF EXISTS api_specifications_tenants;
DROP VIEW IF EXISTS event_api_definitions_tenants;
DROP VIEW IF EXISTS documents_tenants;
DROP VIEW IF EXISTS bundles_tenants;
DROP VIEW IF EXISTS bundle_instance_auths_tenants;
DROP VIEW IF EXISTS api_definitions_tenants;

DROP INDEX tenant_applications_app_id;
DROP INDEX api_definitions_app_id;
DROP INDEX bundle_instance_auths_bundle_id;
DROP INDEX bundle_instance_auths_owner_id;
DROP INDEX bundles_app_id;
DROP INDEX documents_app_id;
DROP INDEX event_api_definitions_app_id;
DROP INDEX api_specifications_tenants_app_id;
DROP INDEX event_specifications_tenants_app_id;
DROP INDEX fetch_request_document_id;
DROP INDEX fetch_request_specification_id;
DROP INDEX labels_tenant_id;
DROP INDEX labels_app_id;
DROP INDEX labels_runtime_id;
DROP INDEX packages_app_id;
DROP INDEX products_app_id;
DROP INDEX tombstones_app_id;
DROP INDEX vendors_app_id;
DROP INDEX webhooks_app_id;
DROP INDEX webhooks_runtime_id;
DROP INDEX system_auths_app_id;
DROP INDEX system_auths_runtime_id;

DROP INDEX tenant_runtimes_runtimes_id;
DROP INDEX runtime_contexts_runtime_id;

ALTER TABLE bundle_instance_auths RENAME COLUMN  owner_id to tenant_id;

ALTER TABLE applications ADD COLUMN  tenant_id UUID;
ALTER TABLE api_definitions ADD  COLUMN  tenant_id UUID;
ALTER TABLE bundles ADD  COLUMN  tenant_id UUID;
ALTER TABLE bundle_references ADD  COLUMN  tenant_id UUID;
ALTER TABLE documents ADD  COLUMN  tenant_id UUID;
ALTER TABLE event_api_definitions ADD  COLUMN  tenant_id UUID;
ALTER TABLE fetch_requests ADD  COLUMN  tenant_id UUID;
ALTER TABLE packages ADD  COLUMN  tenant_id UUID;
ALTER TABLE products ADD  COLUMN  tenant_id UUID;
ALTER TABLE specifications ADD  COLUMN  tenant_id UUID;
ALTER TABLE tombstones ADD  COLUMN  tenant_id UUID;
ALTER TABLE vendors ADD  COLUMN  tenant_id UUID;
ALTER TABLE webhooks ADD  COLUMN  tenant_id UUID;

ALTER TABLE runtimes ADD COLUMN tenant_id UUID;
ALTER TABLE runtime_contexts ADD COLUMN tenant_id UUID;

UPDATE applications a
SET tenant_id = (SELECT tenant_id FROM tenant_applications ta
                 WHERE ta.id = a.id AND
                     (NOT EXISTS(SELECT 1 FROM business_tenant_mappings WHERE parent = ta.tenant_id) -- the tenant has no children
                         OR
                      (NOT EXISTS (SELECT 1 FROM tenant_applications ta2
                                   WHERE ta2.id = a.id AND
                                           tenant_id IN (SELECT id FROM business_tenant_mappings WHERE parent = ta.tenant_id))) -- there is no child that has access
                         ));

UPDATE api_definitions SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);
UPDATE bundles SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);
UPDATE bundle_references SET tenant_id = (SELECT tenant_id FROM bundles WHERE id = bundle_id);
UPDATE documents SET tenant_id = (SELECT tenant_id FROM bundles WHERE id = bundle_id);
UPDATE event_api_definitions SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);
UPDATE specifications SET tenant_id = (SELECT tenant_id FROM api_definitions WHERE id = api_def_id) WHERE api_def_id IS NOT NULL;
UPDATE specifications SET tenant_id = (SELECT tenant_id FROM event_api_definitions WHERE id = event_def_id) WHERE event_def_id IS NOT NULL;
UPDATE fetch_requests SET tenant_id = (SELECT tenant_id FROM documents WHERE id = document_id) WHERE document_id IS NOT NULL;
UPDATE fetch_requests SET tenant_id = (SELECT tenant_id FROM specifications WHERE id = spec_id) WHERE spec_id IS NOT NULL;
UPDATE packages SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);
UPDATE products SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);
UPDATE tombstones SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);
UPDATE vendors SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id);

UPDATE runtimes r
SET tenant_id = (SELECT tenant_id FROM tenant_runtimes tr
                 WHERE tr.id = r.id AND
                     (NOT EXISTS(SELECT 1 FROM business_tenant_mappings WHERE parent = tr.tenant_id) -- the tenant has no children
                         OR
                      (NOT EXISTS (SELECT 1 FROM tenant_runtimes tr2
                                   WHERE tr2.id = r.id AND
                                           tenant_id IN (SELECT id FROM business_tenant_mappings WHERE parent = tr.tenant_id))) -- there is no child that has access
                         ));

UPDATE webhooks SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id) WHERE app_id IS NOT NULL;
UPDATE webhooks SET tenant_id = (SELECT tenant_id FROM runtimes WHERE id = runtime_id) WHERE runtime_id IS NOT NULL;

UPDATE labels SET tenant_id = (SELECT tenant_id FROM applications WHERE id = app_id) WHERE app_id IS NOT NULL;
UPDATE labels SET tenant_id = (SELECT tenant_id FROM runtimes WHERE id = runtime_id) WHERE runtime_id IS NOT NULL;

UPDATE runtime_contexts SET tenant_id = (SELECT tenant_id FROM runtimes WHERE id = runtime_id);

ALTER TABLE api_definitions DROP CONSTRAINT  api_definitions_package_id_fk;
ALTER TABLE api_definitions DROP CONSTRAINT  api_definitions_application_id_fk;
ALTER TABLE bundle_instance_auths DROP CONSTRAINT  bundle_instance_auths_bundle_id_fk;
ALTER TABLE bundle_instance_auths DROP CONSTRAINT  bundle_instance_auths_tenant_id_fk;
ALTER TABLE bundles DROP CONSTRAINT   bundles_application_id_fk;
ALTER TABLE documents DROP CONSTRAINT  documents_application_id_fk;
ALTER TABLE documents DROP CONSTRAINT  documents_bundle_id_fk;
ALTER TABLE event_api_definitions DROP CONSTRAINT  event_api_definitions_package_id_fk ;
ALTER TABLE event_api_definitions DROP CONSTRAINT  event_api_definitions_application_id_fk;
ALTER TABLE fetch_requests DROP CONSTRAINT  fetch_requests_document_id_fk;
ALTER TABLE labels DROP CONSTRAINT  labels_application_id_fk;
ALTER TABLE labels DROP CONSTRAINT  labels_runtime_id_fk;
ALTER TABLE packages DROP CONSTRAINT  packages_application_id_fk;
ALTER TABLE products DROP CONSTRAINT  products_application_id_fk;
ALTER TABLE tombstones DROP CONSTRAINT  tombstones_application_id_fk;
ALTER TABLE vendors DROP CONSTRAINT  vendors_application_id_fk;
ALTER TABLE webhooks DROP CONSTRAINT  webhooks_runtime_id_fk;
ALTER TABLE webhooks DROP CONSTRAINT  webhooks_application_id_fk;
ALTER TABLE system_auths DROP CONSTRAINT  system_auths_application_id_fk;
ALTER TABLE system_auths DROP CONSTRAINT  system_auths_runtime_id_fk;

ALTER TABLE runtime_contexts DROP CONSTRAINT runtime_contexts_runtime_id_fkey;

ALTER TABLE packages ADD CONSTRAINT packages_tenant_id_id_unique UNIQUE (tenant_id, id);
ALTER TABLE applications ADD CONSTRAINT applications_tenant_id_id_unique UNIQUE (tenant_id, id);
ALTER TABLE bundles ADD CONSTRAINT bundles_tenant_id_id_unique UNIQUE (tenant_id, id);
ALTER TABLE documents ADD CONSTRAINT documents_tenant_id_id_unique UNIQUE (tenant_id, id);
ALTER TABLE runtimes ADD CONSTRAINT runtimes_tenant_id_id_unique UNIQUE (tenant_id, id);
ALTER TABLE api_definitions ADD CONSTRAINT api_defs_tenant_id_id_unique UNIQUE (tenant_id, id);
ALTER TABLE event_api_definitions ADD CONSTRAINT event_api_defs_tenant_id_id_unique UNIQUE (tenant_id, id);

ALTER TABLE labels ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE labels DROP CONSTRAINT check_scenario_label_is_tenant_scoped;

DROP INDEX fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx;
CREATE UNIQUE INDEX fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx ON fetch_requests (tenant_id,
                                                                                                 coalesce(document_id, '00000000-0000-0000-0000-000000000000'),
                                                                                                 coalesce(spec_id, '00000000-0000-0000-0000-000000000000'));

ALTER TABLE webhooks DROP CONSTRAINT webhook_owner_id_unique;

ALTER TABLE webhooks ADD CONSTRAINT webhook_owner_id_unique
    CHECK ((app_template_id IS NOT NULL AND tenant_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS NULL)
        OR (app_template_id IS NULL AND tenant_id IS NOT NULL AND app_id IS NOT NULL AND runtime_id IS NULL AND integration_system_id IS NULL)
        OR (app_template_id IS NULL AND tenant_id IS NOT NULL AND app_id IS NULL AND runtime_id IS NOT NULL AND integration_system_id IS NULL)
        OR (app_template_id IS NULL AND tenant_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS  NOT NULL));

ALTER TABLE applications ADD CONSTRAINT application_tenant_id_name_unique UNIQUE (tenant_id, name, system_number);

ALTER TABLE applications ADD CONSTRAINT applications_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE api_definitions ADD  CONSTRAINT  api_definitions_package_id_fk FOREIGN KEY (tenant_id, package_id) REFERENCES packages(tenant_id, id);
ALTER TABLE api_definitions ADD  CONSTRAINT  api_definitions_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings ON DELETE CASCADE;
ALTER TABLE api_definitions ADD  CONSTRAINT  api_definitions_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications(tenant_id, id) ON DELETE CASCADE;
ALTER TABLE bundle_instance_auths ADD  CONSTRAINT  package_instance_auths_tenant_id_fkey FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE bundle_instance_auths ADD  CONSTRAINT  package_instance_auths_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE bundles ADD  CONSTRAINT  packages_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE bundles ADD  CONSTRAINT  packages_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE bundle_references ADD  CONSTRAINT  bundle_references_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE documents ADD  CONSTRAINT  documents_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE documents ADD  CONSTRAINT  documents_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE documents ADD  CONSTRAINT  documents_bundle_id_fk FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE event_api_definitions ADD  CONSTRAINT  event_api_definitions_package_id_fk FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id);
ALTER TABLE event_api_definitions ADD  CONSTRAINT  event_api_definitions_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE event_api_definitions ADD  CONSTRAINT  event_api_definitions_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE fetch_requests ADD  CONSTRAINT  fetch_requests_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE fetch_requests ADD  CONSTRAINT  fetch_requests_tenant_id_fkey2 FOREIGN KEY (tenant_id, document_id) REFERENCES documents (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE labels ADD  CONSTRAINT  labels_tenant_id_fkey  FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE labels ADD  CONSTRAINT  labels_tenant_id_fkey1 FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE packages ADD  CONSTRAINT  packages_application_tenant_fk FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE packages ADD  CONSTRAINT  packages_tenant_id_fkey_cascade FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE products ADD  CONSTRAINT  products_application_tenant_fk FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE products ADD  CONSTRAINT  products_tenant_id_fkey_cascade FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE specifications ADD  CONSTRAINT  specifications_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE tombstones ADD  CONSTRAINT  tombstones_application_tenant_fk FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE tombstones ADD  CONSTRAINT  tombstones_tenant_id_fkey_cascade FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE vendors ADD  CONSTRAINT  vendors_application_tenant_fk FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE vendors ADD  CONSTRAINT  vendors_tenant_id_fkey_cascade FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE webhooks ADD  CONSTRAINT  webhooks_runtime_id_fkey FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE webhooks ADD  CONSTRAINT  webhooks_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE webhooks ADD  CONSTRAINT  webhooks_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE system_auths ADD  CONSTRAINT  system_auths_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE system_auths ADD  CONSTRAINT  system_auths_tenant_id_fkey1 FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE runtimes ADD CONSTRAINT runtimes_tenant_constraint FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE runtime_contexts ADD CONSTRAINT runtime_contexts_tenant_id_fkey FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE runtime_contexts ADD CONSTRAINT runtime_contexts_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_apis;
DROP VIEW IF EXISTS tenants_apps;
DROP VIEW IF EXISTS tenants_bundles;
DROP VIEW IF EXISTS tenants_events;
DROP VIEW IF EXISTS tenants_packages;
DROP VIEW IF EXISTS tenants_products;
DROP VIEW IF EXISTS tenants_vendors;
DROP VIEW IF EXISTS tenants_tombstones;

DROP FUNCTION IF EXISTS apps_subaccounts_func();
DROP FUNCTION IF EXISTS consumers_provider_for_runtimes_func();
DROP FUNCTION IF EXISTS uuid_or_null();

CREATE OR REPLACE FUNCTION apps_subaccounts_func()
    RETURNS TABLE
            (
                id                 uuid,
                tenant_id          text,
                provider_tenant_id text
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT l.app_id                 AS id,
               asa.selector_value::text AS tenant_id,
               asa.selector_value::text AS provider_tenant_id
        FROM labels l
                 -- 2) Get subaccounts in those scenarios (Putting a subaccount in a
                 -- scenario will reflect on creating an ASA for the subaccount.
                 JOIN automatic_scenario_assignments asa
                      ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario::text AND
                         asa.selector_key::text = 'global_subaccount_id'::text
             -- 1) Get all scenario labels for applications
        WHERE l.app_id IS NOT NULL
          AND l.key::text = 'scenarios'::text;
END
$$;

CREATE OR REPLACE FUNCTION consumers_provider_for_runtimes_func()
    RETURNS TABLE
            (
                provider_tenant  text,
                consumer_tenants jsonb
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT l1.value ->> 0 AS provider_tenant, l2.value AS consumer_tenants
        FROM (SELECT *
              FROM labels
              WHERE key::text = 'global_subaccount_id'
                AND (value ->> 0) IS NOT NULL) l1 -- Get the subaccount for each runtime
                 JOIN (SELECT * FROM labels WHERE key::text = 'consumer_subaccount_ids') l2 -- Get all the consumer subaccounts for each runtime
                      ON l1.runtime_id = l2.runtime_id AND l1.runtime_id IS NOT NULL;
END
$$;

CREATE OR REPLACE VIEW tenants_apis
            (tenant_id, provider_tenant_id, id, app_id, name, description, group_name, default_auth, version_value,
             version_deprecated,
             version_deprecated_since, version_for_removal, ord_id, short_description, system_instance_aware,
             api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries,
             labels, package_id, visibility, disabled, part_of_products, line_of_business, industry, ready, created_at,
             updated_at, deleted_at, error, implementation_standard, custom_implementation_standard,
             custom_implementation_standard_description, target_urls, extensible, successors, resource_hash)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                apis.short_description,
                apis.system_instance_aware,
                CASE
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'ODATA'::text THEN 'odata-v2'::text
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'OPEN_API'::text THEN 'rest'::text
                    ELSE apis.api_protocol::text
                    END AS api_protocol,
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
                apis.extensible,
                apis.successors,
                apis.resource_hash
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM applications a1
               UNION ALL
               SELECT a.id,
                      t.parent::text AS tenant_id,
                      t.parent::text AS provider_tenant_id
               FROM applications a
                        JOIN business_tenant_mappings t ON t.id = a.tenant_id
               WHERE t.parent IS NOT NULL
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id, (SELECT id::text FROM business_tenant_mappings WHERE external_tenant = a_s.tenant_id::text), cpr.provider_tenant::text AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? a_s.tenant_id::text) t_apps ON apis.app_id = t_apps.id
         LEFT JOIN specifications specs ON apis.id = specs.api_def_id;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, provider_tenant_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, product_type)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                tmpl.name AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM applications a1
               UNION ALL
               SELECT a.id,
                      t.parent::text AS tenant_id,
                      t.parent::text AS provider_tenant_id
               FROM applications a
                        JOIN business_tenant_mappings t ON t.id = a.tenant_id
               WHERE t.parent IS NOT NULL
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id,(SELECT id::text FROM business_tenant_mappings WHERE external_tenant = a_s.tenant_id::text), cpr.provider_tenant::text AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? a_s.tenant_id::text) t_apps
              ON apps.id = t_apps.id;

CREATE OR REPLACE VIEW tenants_bundles
            (tenant_id, provider_tenant_id, id, app_id, name, description, instance_auth_request_json_schema,
             default_instance_auth, ord_id,
             short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at,
             deleted_at, error, correlation_ids)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
                b.id,
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
                b.correlation_ids
FROM bundles b
         JOIN (SELECT a1.id,
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM applications a1
               UNION ALL
               SELECT a.id,
                      t.parent::text AS tenant_id,
                      t.parent::text AS provider_tenant_id
               FROM applications a
                        JOIN business_tenant_mappings t ON t.id = a.tenant_id
               WHERE t.parent IS NOT NULL
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id, (SELECT id::text FROM business_tenant_mappings WHERE external_tenant = a_s.tenant_id::text), cpr.provider_tenant::text AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? a_s.tenant_id::text) t_apps ON b.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_events
            (tenant_id, provider_tenant_id, id, app_id, name, description, group_name, version_value,
             version_deprecated,
             version_deprecated_since, version_for_removal, ord_id, short_description, system_instance_aware,
             changelog_entries, links, tags, countries, release_status, sunset_date, labels, package_id, visibility,
             disabled, part_of_products, line_of_business, industry, ready, created_at, updated_at, deleted_at, error,
             extensible, successors, resource_hash)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                events.extensible,
                events.successors,
                events.resource_hash
FROM event_api_definitions events
         JOIN (SELECT a1.id,
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM applications a1
               UNION ALL
               SELECT a.id,
                      t.parent::text AS tenant_id,
                      t.parent::text AS provider_tenant_id
               FROM applications a
                        JOIN business_tenant_mappings t ON t.id = a.tenant_id
               WHERE t.parent IS NOT NULL
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id, (SELECT id::text FROM business_tenant_mappings WHERE external_tenant = a_s.tenant_id::text), cpr.provider_tenant::text AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? a_s.tenant_id::text) t_apps ON events.app_id = t_apps.id;

CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, provider_tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type,
             event_spec_format,
             event_spec_type, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                t_api_event_def.provider_tenant_id,
                spec.id,
                spec.api_def_id,
                spec.event_def_id,
                spec.spec_data,
                spec.api_spec_format,
                spec.api_spec_type,
                spec.event_spec_format,
                spec.event_spec_type,
                spec.custom_type,
                spec.created_at
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id::text,
                      a.provider_tenant_id::text
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id::text,
                      e.provider_tenant_id::text
               FROM tenants_events e) t_api_event_def
              ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;

CREATE OR REPLACE FUNCTION get_id_tenant_id_index()
    RETURNS TABLE(table_name TEXT, id UUID, tenant_id UUID) AS
$func$
DECLARE
    compass_table text;
    sql varchar;
BEGIN
    FOR compass_table IN
        SELECT DISTINCT t.table_name
        FROM   information_schema.tables t
                   INNER JOIN information_schema.columns c1 ON t.table_name = c1.table_name
                   INNER JOIN information_schema.columns c2 ON t.table_name = c2.table_name
        WHERE  t.table_schema = 'public'
          AND t.table_type = 'BASE TABLE'
          AND c1.column_name = 'id'
          AND c2.column_name = 'tenant_id'
        LOOP
            sql := 'SELECT ''' || compass_table || '''::TEXT as table_name, id, tenant_id FROM public.' || compass_table || ' WHERE tenant_id IS NOT NULL;';
            RAISE NOTICE 'Executing SQL query: %', sql;
            RETURN QUERY EXECUTE sql;
        END LOOP;
END
$func$  LANGUAGE plpgsql;

-- An index view containing all the entity ids with their owning tenant.
-- It is dynamically updated with the data from each table in the public schema containing 'id' and 'tenant_id' columns.
CREATE MATERIALIZED VIEW id_tenant_id_index AS
SELECT id, tenant_id FROM get_id_tenant_id_index();

CREATE UNIQUE INDEX id_tenant_id_index_unique ON id_tenant_id_index(id);

DROP TRIGGER add_runtime_to_parent_tenants ON tenant_runtimes;
DROP TRIGGER add_app_to_parent_tenants ON tenant_applications;

DROP FUNCTION IF EXISTS insert_parent_chain();

DROP TRIGGER delete_runtime_resource ON tenant_runtimes;
DROP TRIGGER delete_application_resource ON tenant_applications;

DROP FUNCTION IF EXISTS delete_resource();

DROP TRIGGER delete_runtime_from_parent_tenants ON tenant_runtimes;
DROP TRIGGER delete_app_from_parent_tenants ON tenant_applications;

DROP FUNCTION IF EXISTS delete_parent_chain();

DROP TRIGGER delete_tenant ON business_tenant_mappings;

DROP FUNCTION IF EXISTS on_delete_tenant();

DROP TABLE tenant_runtimes;
DROP TABLE tenant_applications;

create table api_runtime_auths(
                                  id         uuid not null
                                      constraint runtime_auths_pkey
                                          primary key
                                      constraint runtime_auths_id_check
                                          check (id <> '00000000-0000-0000-0000-000000000000'::uuid),
                                  tenant_id  uuid not null
                                      constraint api_runtime_auths_tenant_constraint
                                          references business_tenant_mappings
                                          on delete cascade,
                                  runtime_id uuid not null,
                                  api_def_id uuid not null,
                                  value      jsonb,
                                  constraint runtime_auths_tenant_id_fkey1
                                      foreign key (tenant_id, api_def_id) references api_definitions (tenant_id, id)
                                          on delete cascade,
                                  constraint runtime_auths_tenant_id_fkey
                                      foreign key (tenant_id, runtime_id) references runtimes (tenant_id, id)
                                          on update cascade on delete cascade
);

create index runtime_auths_tenant_id_idx
    on api_runtime_auths (tenant_id);

create unique index runtime_auths_tenant_id_runtime_id_api_def_id_idx
    on api_runtime_auths (tenant_id, runtime_id, api_def_id);

create unique index runtime_auths_tenant_id_id_idx
    on api_runtime_auths (tenant_id, id);


COMMIT;
