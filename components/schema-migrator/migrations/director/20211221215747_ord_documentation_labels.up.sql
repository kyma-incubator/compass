BEGIN;

-- packages
ALTER TABLE packages ADD COLUMN documentation_labels JSONB;

-- consumption bundles
ALTER TABLE bundles ADD COLUMN documentation_labels JSONB;

-- api resource
ALTER TABLE api_definitions ADD COLUMN documentation_labels JSONB;

CREATE OR REPLACE VIEW tenants_apis AS
SELECT DISTINCT t_apps.tenant_id AS tenant_id,
                t_apps.provider_tenant_id  AS provider_tenant_id,
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
                apis.resource_hash,
                apis.documentation_labels
FROM api_definitions apis
         JOIN (SELECT a1.id,
                      a1.tenant_id::text,
                      a1.tenant_id::text AS provider_tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT *
               FROM apps_subaccounts_func()
               UNION ALL
               SELECT a_s.id, a_s.tenant_id, (SELECT id::text FROM business_tenant_mappings WHERE external_tenant = cpr.provider_tenant) AS provider_tenant_id
               FROM apps_subaccounts_func() a_s
                        JOIN consumers_provider_for_runtimes_func() cpr
                             ON cpr.consumer_tenants ? (SELECT external_tenant FROM business_tenant_mappings WHERE id = a_s.tenant_id::uuid)) t_apps ON apis.app_id = t_apps.id
         LEFT JOIN specifications specs ON apis.id = specs.api_def_id;

-- event resource
ALTER TABLE event_api_definitions ADD COLUMN documentation_labels JSONB;

-- product
ALTER TABLE products ADD COLUMN documentation_labels JSONB;

-- vendor
ALTER TABLE vendors ADD COLUMN documentation_labels JSONB;

-- system instance
ALTER TABLE applications ADD COLUMN documentation_labels JSONB;

-- ord documentation labels view

CREATE VIEW ord_documentation_labels AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             NULL::uuid     AS bundle_id,
             NULL::uuid     AS application_id,
             NULL::uuid     AS vendor_id,
             NULL::uuid     AS product_id,
             expand.key     AS key,
             elements.value AS value
      FROM packages,
           jsonb_each(packages.documentation_labels) AS expand,
           jsonb_array_elements_text(expand.value) AS elements) AS package_labels
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        NULL::uuid         AS bundle_id,
        NULL::uuid         AS application_id,
        NULL::uuid         AS vendor_id,
        NULL::uuid         AS product_id,
        expand.key         AS key,
        elements.value     AS value
 FROM api_definitions,
      jsonb_each(api_definitions.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        NULL::uuid     AS bundle_id,
        NULL::uuid     AS application_id,
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_each(event_api_definitions.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        id             AS bundle_id,
        NULL::uuid     AS application_id,
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM bundles,
      jsonb_each(bundles.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        NULL::uuid     AS bundle_id,
        id             AS application_id,
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM applications,
      jsonb_each(applications.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        NULL::uuid     AS bundle_id,
        NULL::uuid     AS application_id,
        vendors.id     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM vendors,
      jsonb_each(vendors.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid      AS package_id,
        NULL::uuid      AS api_definition_id,
        NULL::uuid      AS event_definition_id,
        NULL::uuid      AS bundle_id,
        NULL::uuid      AS application_id,
        NULL::uuid      AS vendor_id,
        products.id     AS product_id,
        expand.key      AS key,
        elements.value  AS value
 FROM products,
      jsonb_each(products.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements);


COMMIT;
