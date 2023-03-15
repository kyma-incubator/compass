BEGIN;

DROP VIEW IF EXISTS tenants_apps;

ALTER TABLE products
    ADD COLUMN tags JSONB;

ALTER TABLE vendors
    ADD COLUMN tags JSONB;

ALTER TABLE applications
    ADD COLUMN tags JSONB;

ALTER TABLE bundles
    ADD COLUMN tags JSONB;

ALTER TABLE api_definitions
    ADD COLUMN hierarchy JSONB;

ALTER TABLE api_definitions
    ADD COLUMN supportedUseCases JSONB;

ALTER TABLE event_api_definitions
    ADD COLUMN hierarchy JSONB;

ALTER TABLE event_api_definitions
    ADD COLUMN supportedUseCases JSONB;

CREATE VIEW ord_tags_products AS
SELECT id                  AS product_id,
       elements.value      AS value
FROM products,
     jsonb_array_elements_text(products.tags) AS elements;

CREATE VIEW ord_tags_vendors AS
SELECT id                  AS vendor_id,
       elements.value      AS value
FROM vendors,
     jsonb_array_elements_text(vendors.tags) AS elements;

CREATE VIEW ord_tags_applications AS
SELECT id                  AS application_id,
       elements.value      AS value
FROM applications,
     jsonb_array_elements_text(applications.tags) AS elements;

CREATE VIEW ord_tags_bundles AS
SELECT id                  AS bundle_id,
       elements.value      AS value
FROM bundles,
     jsonb_array_elements_text(bundles.tags) AS elements;

CREATE VIEW ord_hierarchy_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.hierarchy) AS elements;

CREATE VIEW ord_hierarchy_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.hierarchy) AS elements;

CREATE VIEW ord_supported_use_cases_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.supportedUseCases) AS elements;

CREATE VIEW ord_supported_use_cases_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.supportedUseCases) AS elements;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, formation_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, application_namespace, product_type)
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
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                coalesce(apps.application_namespace, tmpl.application_namespace),
                tmpl.name AS product_type
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      apps_subaccounts.formation_id
               FROM apps_subaccounts
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps
              ON apps.id = t_apps.id;

COMMIT;