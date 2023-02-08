BEGIN;

------------------ New Types --------------------------------

CREATE TYPE vendor_type AS ENUM ('sap', 'sap-partner','client_registration');

CREATE TABLE vendors
(
    ord_id    VARCHAR(256) PRIMARY KEY,
    tenant_id UUID         NOT NULL,
    app_id    UUID         NOT NULL,
    CONSTRAINT vendors_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id),
    title     VARCHAR(256) NOT NULL,
    type      vendor_type  NOT NULL,
    labels    JSONB
);

CREATE INDEX ON vendors (tenant_id);
CREATE UNIQUE INDEX ON vendors (tenant_id, ord_id);

CREATE TABLE products
(
    ord_id             VARCHAR(256) PRIMARY KEY,
    tenant_id          UUID         NOT NULL,
    app_id             UUID         NOT NULL,
    CONSTRAINT products_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id),
    title              VARCHAR(256) NOT NULL,
    short_description  VARCHAR(255) NOT NULL,
    vendor             VARCHAR(256) NOT NULL,
    parent             VARCHAR(256),
    sap_ppms_object_id VARCHAR(256),
    labels             JSONB
);

CREATE INDEX ON products (tenant_id);
CREATE UNIQUE INDEX ON products (tenant_id, ord_id);

CREATE TABLE tombstones
(
    ord_id       VARCHAR(256) PRIMARY KEY,
    tenant_id    UUID         NOT NULL,
    app_id       UUID         NOT NULL,
    CONSTRAINT tombstones_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id),
    removal_date VARCHAR(256) NOT NULL
);
------------------------------------------------------------
------------------------- Packages -------------------------

CREATE TYPE policy_level AS ENUM ('sap', 'sap-partner','custom');

DROP VIEW providers;
ALTER TABLE packages
    DROP COLUMN provider;

ALTER TABLE packages
    ADD COLUMN policy_level policy_level NOT NULL,
    ADD COLUMN app_id UUID NOT NULL,
    ADD COLUMN custom_policy_level VARCHAR (256),
    ADD COLUMN vendor VARCHAR (256),
    ADD COLUMN part_of_products JSONB NOT NULL,
    ADD COLUMN line_of_business JSONB,
    ADD COLUMN industry JSONB;

ALTER TABLE packages
    ADD CONSTRAINT packages_vendor_fk
        FOREIGN KEY (tenant_id, vendor) REFERENCES vendors (tenant_id, ord_id);

ALTER TABLE packages
    ADD CONSTRAINT packages_apps_fk
        FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id);

CREATE VIEW package_product AS
SELECT packages.id    AS package_id,
       elements.value AS product_id
FROM packages,
     jsonb_array_elements_text(packages.part_of_products) AS elements;

------------------------------------------------------------
------------------------- Bundles --------------------------

DROP VIEW credential_request_strategies;

ALTER TABLE bundles
    RENAME credential_request_strategies TO credential_exchange_strategies;

CREATE VIEW credential_exchange_strategies AS
SELECT id                                                AS bundle_id,
       credential_request_strategies.type                AS type,
       credential_request_strategies."customType"        AS custom_type,
       credential_request_strategies."customDescription" AS custom_description,
       credential_request_strategies."callbackUrl"       AS callback_url
FROM bundles,
     jsonb_to_recordset(bundles.credential_exchange_strategies) AS credential_request_strategies(type TEXT,
                                                                                                 "customType" TEXT,
                                                                                                 "customDescription" TEXT,
                                                                                                 "callbackUrl" TEXT);

------------------------------------------------------------
------------ Applications - System Instances ---------------

ALTER TABLE applications
    ADD COLUMN base_url VARCHAR(512),
    ADD COLUMN labels   JSONB;

------------------------------------------------------------
--------------------------- APIs ---------------------------
CREATE TYPE visibility AS ENUM ('public', 'internal','private');

ALTER TABLE api_definitions
    ADD COLUMN visibility       visibility, -- ORD Required, Nullable due to backwards compatibility.
    ADD COLUMN disabled         BOOLEAN,
    ADD COLUMN part_of_products JSONB,
    ADD COLUMN line_of_business JSONB,
    ADD COLUMN industry         JSONB;

CREATE VIEW api_product AS
SELECT api_definitions.id AS api_definition_id,
       elements.value     AS product_id
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.part_of_products) AS elements;
------------------------------------------------------------
-------------------------- Events --------------------------

ALTER TABLE event_api_definitions
    ADD COLUMN visibility       visibility, -- ORD Required, Nullable due to backwards compatibility.
    ADD COLUMN disabled         BOOLEAN,
    ADD COLUMN part_of_products JSONB,
    ADD COLUMN line_of_business JSONB,
    ADD COLUMN industry         JSONB;

CREATE VIEW event_product AS
SELECT event_api_definitions.id AS event_definition_id,
       elements.value           AS product_id
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.part_of_products) AS elements;

-------------------------------------------------------------------
-------------------------- Generic Views --------------------------

CREATE VIEW line_of_businesses AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             elements.value AS value
      FROM packages,
           jsonb_array_elements_text(packages.line_of_business) AS elements) AS package_lb
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        elements.value     AS value
 FROM api_definitions,
      jsonb_array_elements_text(api_definitions.line_of_business) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_array_elements_text(event_api_definitions.line_of_business) AS elements);


CREATE VIEW industries AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             elements.value AS value
      FROM packages,
           jsonb_array_elements_text(packages.industry) AS elements) AS package_industries
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        elements.value     AS value
 FROM api_definitions,
      jsonb_array_elements_text(api_definitions.industry) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_array_elements_text(event_api_definitions.industry) AS elements);


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

COMMIT;
