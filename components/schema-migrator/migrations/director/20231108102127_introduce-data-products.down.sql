BEGIN;

-- Add runtimeRestriction to packages
ALTER TABLE packages 
       DROP COLUMN runtime_restriction VARCHAR(256);

DROP VIEW IF EXISTS tenants_packages;

CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, formation_id, id, ord_id, title, short_description, description, version, package_links,
             links, licence_type, tags, countries, labels, policy_level, app_id, custom_policy_level, vendor,
             part_of_products, line_of_business, industry, resource_hash, support_info)
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
              ON p.app_id = t_apps.id;


COMMIT;
