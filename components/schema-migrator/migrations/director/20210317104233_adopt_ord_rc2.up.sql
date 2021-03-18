BEGIN;

ALTER TABLE api_definitions
    ADD COLUMN implementation_standard                        VARCHAR(256),
    ADD COLUMN custom_implementation_standard                 VARCHAR(256),
    ADD COLUMN custom_implementation_standard_description     VARCHAR(255);

ALTER TABLE products
    DROP COLUMN sap_ppms_object_id;
ALTER TABLE products
    ADD COLUMN correlation_ids JSONB;

ALTER TABLE applications
    ADD COLUMN correlation_ids JSONB;

ALTER TABLE vendors
    DROP COLUMN type;
ALTER TABLE vendors
    ADD COLUMN sap_partner BOOLEAN;

CREATE VIEW correlation_ids AS
SELECT *
FROM (SELECT applications.id            AS application_id,
             NULL::varchar(256)         AS product_id,
             elements.value             AS value
      FROM applications,
          jsonb_array_elements_text(applications.correlation_ids) AS elements) AS app_correlation_ids
UNION ALL
(SELECT NULL::uuid         AS application_id,
        products.ord_id    AS product_id,
        elements.value     AS value
 FROM products,
     jsonb_array_elements_text(products.correlation_ids) AS elements);

COMMIT;
