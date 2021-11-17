BEGIN;

ALTER TABLE bundles ADD COLUMN correlation_ids JSONB;

DROP VIEW IF EXISTS tenants_bundles;
DROP VIEW IF EXISTS correlation_ids;

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


CREATE VIEW correlation_ids AS
SELECT *
FROM (SELECT applications.id            AS application_id,
             NULL::uuid                 AS product_id,
             NULL::uuid                 AS bundle_id,
             elements.value             AS value
      FROM applications,
           jsonb_array_elements_text(applications.correlation_ids) AS elements) AS app_correlation_ids
UNION ALL
(SELECT NULL::uuid         AS application_id,
        products.id        AS product_id,
        NULL::uuid         AS bundle_id,
        elements.value     AS value
 FROM products,
      jsonb_array_elements_text(products.correlation_ids) AS elements)
UNION ALL
(SELECT NULL::uuid        AS application_id,
        NULL::uuid        AS product_id,
        bundle.id         AS bundle_id,
        elements.value    AS value
 FROM bundles,
      jsonb_array_elements_text(bundles.correlation_ids) AS elements);

COMMIT;
