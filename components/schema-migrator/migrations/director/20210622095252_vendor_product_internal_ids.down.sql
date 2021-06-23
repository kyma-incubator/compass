BEGIN;

--- views ---

DROP VIEW ord_labels;

CREATE VIEW ord_labels AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             NULL::uuid     AS bundle_id,
             NULL::uuid     AS application_id,
             NULL           AS vendor_id,
             NULL           AS product_id,
             expand.key     AS key,
             elements.value AS value
      FROM packages,
           jsonb_each(packages.labels) AS expand,
           jsonb_array_elements_text(expand.value) AS elements) AS package_labels
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        NULL::uuid         AS bundle_id,
        NULL::uuid         AS application_id,
        NULL               AS vendor_id,
        NULL               AS product_id,
        expand.key         AS key,
        elements.value     AS value
 FROM api_definitions,
      jsonb_each(api_definitions.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        NULL::uuid     AS bundle_id,
        NULL::uuid     AS application_id,
        NULL           AS vendor_id,
        NULL           AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_each(event_api_definitions.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        id             AS bundle_id,
        NULL::uuid     AS application_id,
        NULL           AS vendor_id,
        NULL           AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM bundles,
      jsonb_each(bundles.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        NULL::uuid     AS bundle_id,
        id             AS application_id,
        NULL           AS vendor_id,
        NULL           AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM applications,
      jsonb_each(applications.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        NULL::uuid     AS bundle_id,
        NULL::uuid     AS application_id,
        vendors.ord_id AS vendor_id,
        NULL           AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM vendors,
      jsonb_each(vendors.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid      AS package_id,
        NULL::uuid      AS api_definition_id,
        NULL::uuid      AS event_definition_id,
        NULL::uuid      AS bundle_id,
        NULL::uuid      AS application_id,
        NULL            AS vendor_id,
        products.ord_id AS product_id,
        expand.key      AS key,
        elements.value  AS value
 FROM products,
      jsonb_each(products.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements);

---

DROP VIEW partners;

CREATE VIEW partners AS
SELECT vendors.ord_id AS vendor_id,
       elements.value AS value
FROM vendors,
     jsonb_array_elements_text(vendors.partners) AS elements;

---

DROP VIEW correlation_ids;

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

---

DROP VIEW package_product;

CREATE VIEW package_product AS
SELECT packages.id     AS package_id,
       elements.value  AS product_id
FROM packages,
     jsonb_array_elements_text(packages.part_of_products) AS elements;

---

DROP VIEW api_product;

CREATE VIEW api_product AS
SELECT api_definitions.id     AS api_definition_id,
       elements.value         AS product_id
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.part_of_products) AS elements;

---

DROP VIEW event_product;

CREATE VIEW event_product AS
SELECT event_api_definitions.id     AS event_definition_id,
       elements.value               AS product_id
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.part_of_products) AS elements;

--- vendors ---

ALTER TABLE vendors DROP CONSTRAINT vendors_pkey;
ALTER TABLE vendors ADD CONSTRAINT vendors_pkey PRIMARY KEY (ord_id);

ALTER TABLE vendors
    DROP COLUMN id;

--- products ---

ALTER TABLE products DROP CONSTRAINT products_pkey;
ALTER TABLE products ADD CONSTRAINT products_pkey PRIMARY KEY (ord_id);

ALTER TABLE products
    DROP COLUMN id;

--- tombstones ---

ALTER TABLE tombstones DROP CONSTRAINT tombstones_pkey;
ALTER TABLE tombstones ADD CONSTRAINT tombstones_pkey PRIMARY KEY (ord_id);

ALTER TABLE tombstones
    DROP COLUMN id;

COMMIT;
