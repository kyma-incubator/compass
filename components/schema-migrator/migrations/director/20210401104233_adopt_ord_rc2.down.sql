BEGIN;

ALTER TABLE api_definitions
    DROP COLUMN implementation_standard,
    DROP COLUMN custom_implementation_standard,
    DROP COLUMN custom_implementation_standard_description;

ALTER TABLE products
    ADD COLUMN sap_ppms_object_id VARCHAR(256);

ALTER TABLE products
    DROP COLUMN correlation_ids;

COMMIT;
