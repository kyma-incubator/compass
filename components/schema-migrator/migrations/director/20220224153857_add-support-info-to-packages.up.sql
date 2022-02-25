BEGIN;

ALTER TABLE packages
    ADD COLUMN support_info VARCHAR(256);

DROP VIEW IF EXISTS tenants_packages;

-----

CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, provider_tenant_id, id, ord_id, title, short_description, description, version, package_links,
             links, licence_type, tags, countries, labels, policy_level, app_id, custom_policy_level, vendor,
             part_of_products, line_of_business, industry, resource_hash, support_info)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.provider_tenant_id,
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
                      a1.tenant_id::text AS tenant_id,
                      a1.tenant_id::text AS provider_tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id::text,
                      apps_subaccounts_func.provider_tenant_id::text
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id, provider_tenant_id)
               UNION ALL
               SELECT ta.id AS app_id, ta.tenant_id::text AS consumer_tenant, tenant_runtimes.tenant_id::text AS provider_tenant
               FROM (SELECT labels.runtime_id, v ->> 0 AS consumer_tenant
                     FROM labels
                              JOIN jsonb_array_elements(labels.value) AS v ON TRUE
                     WHERE key = 'consumer_subaccount_ids') AS t_rts -- Get runtime and external consumer IDs pairs
                        JOIN business_tenant_mappings t ON t_rts.consumer_tenant = t.external_tenant -- Get runtime and internal consumer IDs pairs
                        JOIN apps_subaccounts_func() ta ON t.id = ta.tenant_id -- Get applications for consumer tenants
                        JOIN tenant_runtimes ON t_rts.runtime_id = tenant_runtimes.id) t_apps
              ON p.app_id = t_apps.id;

-----

COMMIT;
