BEGIN;

DROP VIEW IF EXISTS ord_tags_products;
DROP VIEW IF EXISTS ord_tags_vendors;
DROP VIEW IF EXISTS ord_tags_applications;
DROP VIEW IF EXISTS ord_tags_bundles;

ALTER TABLE products
    DROP COLUMN tags;

ALTER TABLE vendors
    DROP COLUMN tags;

ALTER TABLE applications
    DROP COLUMN tags;

ALTER TABLE bundles
    DROP COLUMN tags;

COMMIT;