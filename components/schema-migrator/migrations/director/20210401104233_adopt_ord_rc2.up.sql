BEGIN;

ALTER TABLE api_definitions
    ADD COLUMN implementation_standard                        VARCHAR(256),
    ADD COLUMN custom_implementation_standard                 VARCHAR(256),
    ADD COLUMN custom_implementation_standard_description     VARCHAR(255);

ALTER TABLE products
    DROP COLUMN sap_ppms_object_id;

ALTER TABLE products
    ADD COLUMN correlation_ids JSONB;

COMMIT;
