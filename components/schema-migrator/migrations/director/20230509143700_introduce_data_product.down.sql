BEGIN;

DROP VIEW IF EXISTS input_port_api_reference;
DROP VIEW IF EXISTS output_port_api_reference;
DROP VIEW IF EXISTS input_port_event_reference;
DROP VIEW IF EXISTS output_port_event_reference;
DROP VIEW IF EXISTS data_product_tenants;
DROP VIEW IF EXISTS tenants_data_products;
DROP VIEW IF EXISTS ord_data_products_tags;
DROP VIEW IF EXISTS ord_data_products_industry;
DROP VIEW IF EXISTS ord_data_products_line_of_business;
DROP VIEW IF EXISTS tenants_input_ports;
DROP VIEW IF EXISTS tenants_output_ports;

ALTER TABLE bundle_references
    DROP CONSTRAINT IF EXISTS bundle_references_data_product_id_fkey;

ALTER TABLE bundle_references
    DROP COLUMN IF EXISTS data_product_id;

ALTER TABLE bundle_references
    DROP CONSTRAINT IF EXISTS valid_refs;

DROP TABLE IF EXISTS port_api_reference;
DROP TABLE IF EXISTS port_event_reference;
DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS data_products;

COMMIT;
