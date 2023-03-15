BEGIN;

DROP VIEW IF EXISTS ord_tags_products;
DROP VIEW IF EXISTS ord_tags_vendors;
DROP VIEW IF EXISTS ord_tags_applications;
DROP VIEW IF EXISTS ord_tags_bundles;
DROP VIEW IF EXISTS ord_hierarchy_event_definitions;
DROP VIEW IF EXISTS ord_hierarchy_api_definitions;
DROP VIEW IF EXISTS ord_supported_use_cases_event_definitions;
DROP VIEW IF EXISTS ord_supported_use_cases_api_definitions;
DROP VIEW IF EXISTS tenants_apps;

ALTER TABLE products
    DROP COLUMN tags;

ALTER TABLE vendors
    DROP COLUMN tags;

ALTER TABLE applications
    DROP COLUMN tags;

ALTER TABLE bundles
    DROP COLUMN tags;

ALTER TABLE api_definitions
    DROP COLUMN hierarchy;

ALTER TABLE api_definitions
    DROP COLUMN supportedUseCases;

ALTER TABLE event_api_definitions
    DROP COLUMN hierarchy;

ALTER TABLE event_api_definitions
    DROP COLUMN supportedUseCases;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, formation_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error,
             app_template_id, correlation_ids, system_number, product_type)
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