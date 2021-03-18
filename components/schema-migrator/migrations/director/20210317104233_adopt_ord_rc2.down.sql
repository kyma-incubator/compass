BEGIN;

DROP VIEW correlation_ids;

ALTER TABLE api_definitions
    DROP COLUMN implementation_standard,
    DROP COLUMN custom_implementation_standard,
    DROP COLUMN custom_implementation_standard_description;

ALTER TABLE products
    ADD COLUMN sap_ppms_object_id VARCHAR(256);

ALTER TABLE products
    DROP COLUMN correlation_ids;

ALTER TABLE applications
    DROP COLUMN correlation_ids;

ALTER TABLE vendors
    ADD COLUMN type VARCHAR(256) NOT NULL;

ALTER TABLE vendors
    DROP COLUMN sap_partner;

COMMIT;
