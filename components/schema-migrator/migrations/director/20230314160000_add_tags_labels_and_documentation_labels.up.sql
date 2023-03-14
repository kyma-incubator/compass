BEGIN;

ALTER TABLE products
    ADD COLUMN tags JSONB;

ALTER TABLE vendors
    ADD COLUMN tags JSONB;

ALTER TABLE applications
    ADD COLUMN tags JSONB;

ALTER TABLE bundles
    ADD COLUMN tags JSONB;

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

COMMIT;