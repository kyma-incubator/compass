BEGIN;

DROP VIEW ord_labels;
DROP VIEW industries;
DROP VIEW line_of_businesses;

DROP VIEW event_product;

ALTER TABLE event_api_definitions
    DROP COLUMN visibility,
    DROP COLUMN disabled,
    DROP COLUMN part_of_products,
    DROP COLUMN line_of_business,
    DROP COLUMN industry;

DROP VIEW api_product;

ALTER TABLE api_definitions
    DROP COLUMN visibility,
    DROP COLUMN disabled,
    DROP COLUMN part_of_products,
    DROP COLUMN line_of_business,
    DROP COLUMN industry;

DROP TYPE visibility;

ALTER TABLE applications
    DROP COLUMN base_url,
    DROP COLUMN labels;

DROP VIEW credential_exchange_strategies;

ALTER TABLE bundles
    RENAME credential_exchange_strategies TO credential_request_strategies;

CREATE VIEW credential_request_strategies AS
SELECT id                                          AS bundle_id,
       credential_request_strategies.type          AS type,
       credential_request_strategies."callbackUrl" AS callback_url
FROM bundles,
     jsonb_to_recordset(bundles.credential_request_strategies) AS credential_request_strategies(type TEXT, "callbackUrl" TEXT);

DROP VIEW package_product;

ALTER TABLE packages
    DROP CONSTRAINT packages_vendor_fk;

ALTER TABLE packages
    DROP CONSTRAINT packages_apps_fk;

ALTER TABLE packages
    DROP COLUMN policy_level,
    DROP COLUMN custom_policy_level,
    DROP COLUMN vendor,
    DROP COLUMN part_of_products,
    DROP COLUMN line_of_business,
    DROP COLUMN industry;

ALTER TABLE packages
    ADD COLUMN provider JSONB;

CREATE VIEW providers AS
SELECT packages.id                        AS package_id,
       packages.provider ->> 'name'       AS name,
       packages.provider ->> 'department' AS department
FROM packages;

DROP TYPE policy_level;

DROP TABLE tombstones;
DROP TABLE products;
DROP TABLE vendors;

DROP TYPE vendor_type;

CREATE VIEW ord_labels AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             NULL::uuid     AS bundle_id,
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
        expand.key     AS key,
        elements.value AS value
 FROM bundles,
      jsonb_each(bundles.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements);

COMMIT;
