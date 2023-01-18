BEGIN;

-- We need to drop these views because they references the 'policy_level' type. After we modify the `policy_level` type we can recreate the views as is done eventually in this script.
DROP VIEW IF EXISTS tenants_packages;
DROP VIEW IF EXISTS packages_tenants;

---

ALTER TYPE policy_level RENAME TO policy_level_old;

CREATE TYPE policy_level AS ENUM ('custom', 'sap:core:v1','sap:partner:v1');

ALTER TABLE packages
    ALTER COLUMN policy_level TYPE policy_level
        USING policy_level::text::policy_level;

DROP TYPE policy_level_old;

---

CREATE OR REPLACE VIEW tenants_packages
            (tenant_id, id, ord_id, title, short_description, description, version, package_links,
             links, licence_type, tags, countries, labels, policy_level, app_id, custom_policy_level, vendor,
             part_of_products, line_of_business, industry, resource_hash, support_info)
AS
SELECT DISTINCT t_apps.tenant_id,
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
                      a1.tenant_id AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id
               FROM apps_subaccounts) t_apps
              ON p.app_id = t_apps.id;

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

COMMIT;
