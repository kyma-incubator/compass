BEGIN;

DROP VIEW IF EXISTS tenants_vendors;

CREATE OR REPLACE VIEW tenants_vendors
            (tenant_id, formation_id, ord_id, app_id, title, labels, tags, partners, id, documentation_labels)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                v.ord_id,
                coalesce(v.app_id, t_apps.id),
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

-----

DROP VIEW IF EXISTS tenants_products;

CREATE OR REPLACE VIEW tenants_products
            (tenant_id, formation_id, ord_id, app_id, title, short_description, vendor, parent, labels, tags,
             correlation_ids, id, documentation_labels, description)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                p.ord_id,
                coalesce(p.app_id, t_apps.id),
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


COMMIT;
