BEGIN;

ALTER TABLE vendors ALTER COLUMN app_id DROP NOT NULL;
ALTER TABLE products ALTER COLUMN app_id DROP NOT NULL;

ALTER TABLE products DROP CONSTRAINT products_vendor_fk;
ALTER TABLE packages DROP CONSTRAINT packages_vendor_fk;

ALTER TABLE vendors DROP CONSTRAINT vendor_ord_id_unique;
ALTER TABLE vendors ADD CONSTRAINT vendor_ord_id_unique UNIQUE (ord_id);

ALTER TABLE products DROP CONSTRAINT product_ord_id_unique;
ALTER TABLE products ADD CONSTRAINT product_ord_id_unique UNIQUE (ord_id);

ALTER TABLE products ADD CONSTRAINT products_vendor_fk FOREIGN KEY (vendor) REFERENCES vendors(ord_id);
ALTER TABLE packages ADD CONSTRAINT packages_vendor_fk FOREIGN KEY (vendor) REFERENCES vendors(ord_id);

------

DROP VIEW api_product;

CREATE VIEW api_product AS
SELECT api_definitions.id     AS api_definition_id,
       p.id                   AS product_id
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = api_definitions.app_id OR p.app_id IS NULL;

------

DROP VIEW event_product;

CREATE VIEW event_product AS
SELECT event_api_definitions.id     AS event_definition_id,
       p.id                         AS product_id
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = event_api_definitions.app_id OR p.app_id IS NULL;

------

DROP VIEW package_product;

CREATE VIEW package_product AS
SELECT packages.id     AS package_id,
       p.id            AS product_id
FROM packages,
     jsonb_array_elements_text(packages.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = packages.app_id OR p.app_id IS NULL;

------

DROP VIEW IF EXISTS tenants_vendors;

CREATE OR REPLACE VIEW tenants_vendors AS
SELECT DISTINCT t_apps.tenant_id AS tenant_id,
                t_apps.provider_tenant_id AS provider_tenant_id,
                v.*
FROM vendors v
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
                             ON cpr.consumer_tenants ? (SELECT external_tenant FROM business_tenant_mappings WHERE id = a_s.tenant_id::uuid)) t_apps ON v.app_id = t_apps.id OR v.app_id IS NULL;

------

DROP VIEW IF EXISTS tenants_products;

CREATE OR REPLACE VIEW tenants_products AS
SELECT DISTINCT t_apps.tenant_id AS tenant_id,
                t_apps.provider_tenant_id AS provider_tenant_id,
                p.*
FROM products p
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
                             ON cpr.consumer_tenants ? (SELECT external_tenant FROM business_tenant_mappings WHERE id = a_s.tenant_id::uuid)) t_apps ON p.app_id = t_apps.id OR p.app_id IS NULL;

COMMIT;
