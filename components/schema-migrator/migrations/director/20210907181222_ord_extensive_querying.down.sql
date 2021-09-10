BEGIN;

DROP VIEW tenants_specifications;
DROP VIEW tenants_apis;
DROP VIEW api_product;
DROP VIEW event_product;
DROP VIEW package_product;

CREATE OR REPLACE VIEW tenants_apis  AS
SELECT DISTINCT t_apps.tenant_id, apis.id, apis.app_id, apis.name, apis.description, apis.group_name, apis.default_auth,
                apis.version_value, apis.version_deprecated, apis.version_deprecated_since, apis.version_for_removal,
                apis.ord_id, apis.short_description, apis.system_instance_aware,
                CASE
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'ODATA' THEN 'odata-v2'::api_protocol
                    WHEN apis.api_protocol IS NULL AND specs.api_spec_type::text = 'OPEN_API' THEN 'rest'::api_protocol
                    ELSE apis.api_protocol
                    END                                            AS api_protocol,
                apis.tags, apis.countries, apis.links, apis.api_resource_links, apis.release_status, apis.sunset_date,
                apis.changelog_entries, apis.labels, apis.package_id, apis.visibility, apis.disabled, apis.part_of_products,
                apis.line_of_business, apis.industry, apis.ready, apis.created_at, apis.updated_at, apis.deleted_at, apis.error,
                apis.implementation_standard, apis.custom_implementation_standard, apis.custom_implementation_standard_description,
                apis.target_urls, apis.extensible, apis.successors, apis.resource_hash
FROM api_definitions AS apis
         INNER JOIN (
--  select GAs
    SELECT a1.id, a1.tenant_id FROM applications AS a1
    UNION ALL
--  select CRM
    SELECT a.id, t.parent as tenant_id FROM applications AS a
                                                INNER JOIN  business_tenant_mappings AS t ON t.id = a.tenant_id WHERE t.parent IS NOT NULL
    UNION ALL
--  select SA
    SELECT l.app_id as id, asa.selector_value::uuid as tenant_id FROM labels AS l
                                                                          INNER JOIN automatic_scenario_assignments AS asa ON asa.tenant_id = l.tenant_id AND l.value ? asa.scenario AND asa.selector_key='global_subaccount_id'
    WHERE l.app_id IS NOT NULL AND l.key = 'scenarios') AS t_apps ON apis.app_id = t_apps.id
         LEFT JOIN specifications as specs on apis.id = specs.api_def_id;

CREATE OR REPLACE VIEW tenants_specifications  AS
SELECT DISTINCT t_api_event_def.tenant_id, spec.id, spec.api_def_id, spec.event_def_id, spec.spec_data, spec.api_spec_format, spec.api_spec_type, spec.event_spec_format, spec.event_spec_type, spec.custom_type, spec.created_at
FROM specifications AS spec
         INNER JOIN (
--  select APIs
    SELECT a.id, a.tenant_id FROM tenants_apis AS a
    UNION ALL
--  select APIs
    SELECT e.id, e.tenant_id FROM tenants_events AS e
) AS t_api_event_def ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;

---

CREATE VIEW api_product AS
SELECT api_definitions.id     AS api_definition_id,
       api_definitions.app_id AS app_id,
       elements.value         AS product_id
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.part_of_products) AS elements;

---

CREATE VIEW event_product AS
SELECT event_api_definitions.id     AS event_definition_id,
       event_api_definitions.app_id AS app_id,
       elements.value               AS product_id
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.part_of_products) AS elements;

---

CREATE VIEW package_product AS
SELECT packages.id     AS package_id,
       packages.app_id AS app_id,
       elements.value  AS product_id
FROM packages,
     jsonb_array_elements_text(packages.part_of_products) AS elements;

---

ALTER TABLE bundle_references
DROP CONSTRAINT bundle_references_pk;

ALTER TABLE bundle_references
DROP COLUMN id;

COMMIT;
