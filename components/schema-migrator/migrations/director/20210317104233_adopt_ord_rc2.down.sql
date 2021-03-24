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


CREATE TABLE convert_json_to_string_temp (
    id UUID,
    url VARCHAR(256)
);

INSERT INTO convert_json_to_string_temp
    (SELECT id, jsonb_array_elements_text(target_urls) FROM api_definitions);

ALTER TABLE api_definitions
    ADD COLUMN  target_url VARCHAR(256) NOT NULL DEFAULT '';


UPDATE api_definitions
SET target_url = convert_json_to_string_temp.url
FROM convert_json_to_string_temp
WHERE api_definitions.id = convert_json_to_string_temp.id;

DROP TABLE convert_json_to_string_temp;

ALTER TABLE api_definitions
    DROP COLUMN target_urls;

COMMIT;
