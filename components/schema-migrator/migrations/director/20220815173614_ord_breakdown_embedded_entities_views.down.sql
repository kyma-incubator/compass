BEGIN;

DROP VIEW IF EXISTS ord_labels_packages;
DROP VIEW IF EXISTS ord_labels_api_definitions;
DROP VIEW IF EXISTS ord_labels_event_definitions;
DROP VIEW IF EXISTS ord_labels_bundles;
DROP VIEW IF EXISTS ord_labels_applications;
DROP VIEW IF EXISTS ord_labels_vendors;
DROP VIEW IF EXISTS ord_labels_products;

DROP VIEW IF EXISTS ord_documentation_labels_packages;
DROP VIEW IF EXISTS ord_documentation_labels_api_definitions;
DROP VIEW IF EXISTS ord_documentation_labels_event_definitions;
DROP VIEW IF EXISTS ord_documentation_labels_bundles;
DROP VIEW IF EXISTS ord_documentation_labels_applications;
DROP VIEW IF EXISTS ord_documentation_labels_vendors;
DROP VIEW IF EXISTS ord_documentation_labels_products;

DROP VIEW IF EXISTS line_of_businesses_packages;
DROP VIEW IF EXISTS line_of_businesses_api_definitions;
DROP VIEW IF EXISTS line_of_businesses_event_definitions;

DROP VIEW IF EXISTS countries_packages;
DROP VIEW IF EXISTS countries_api_definitions;
DROP VIEW IF EXISTS countries_event_definitions;

DROP VIEW IF EXISTS changelog_entries_api_definitions;
DROP VIEW IF EXISTS changelog_entries_event_definitions;

DROP VIEW IF EXISTS tags_packages;
DROP VIEW IF EXISTS tags_api_definitions;
DROP VIEW IF EXISTS tags_event_definitions;

DROP VIEW IF EXISTS industries_packages;
DROP VIEW IF EXISTS industries_api_definitions;
DROP VIEW IF EXISTS industries_event_definitions;

DROP VIEW IF EXISTS links_packages;
DROP VIEW IF EXISTS links_api_definitions;
DROP VIEW IF EXISTS links_event_definitions;
DROP VIEW IF EXISTS links_bundles;

DROP VIEW IF EXISTS correlation_ids_applications;
DROP VIEW IF EXISTS correlation_ids_products;
DROP VIEW IF EXISTS correlation_ids_bundles;

---

CREATE VIEW ord_labels AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             NULL::uuid     AS bundle_id,
             NULL::uuid     AS application_id,
             NULL::uuid     AS vendor_id,
             NULL::uuid     AS product_id,
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
        NULL::uuid         AS vendor_id,
        NULL::uuid         AS product_id,
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
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
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
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
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
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
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
        vendors.id     AS vendor_id,
        NULL::uuid     AS product_id,
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
        NULL::uuid      AS vendor_id,
        products.id     AS product_id,
        expand.key      AS key,
        elements.value  AS value
 FROM products,
      jsonb_each(products.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements);

---

CREATE VIEW ord_documentation_labels AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             NULL::uuid     AS bundle_id,
             NULL::uuid     AS application_id,
             NULL::uuid     AS vendor_id,
             NULL::uuid     AS product_id,
             expand.key     AS key,
             elements.value AS value
      FROM packages,
           jsonb_each(packages.documentation_labels) AS expand,
           jsonb_array_elements_text(expand.value) AS elements) AS package_labels
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        NULL::uuid         AS bundle_id,
        NULL::uuid         AS application_id,
        NULL::uuid         AS vendor_id,
        NULL::uuid         AS product_id,
        expand.key         AS key,
        elements.value     AS value
 FROM api_definitions,
      jsonb_each(api_definitions.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        NULL::uuid     AS bundle_id,
        NULL::uuid     AS application_id,
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_each(event_api_definitions.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        id             AS bundle_id,
        NULL::uuid     AS application_id,
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM bundles,
      jsonb_each(bundles.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        NULL::uuid     AS bundle_id,
        id             AS application_id,
        NULL::uuid     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM applications,
      jsonb_each(applications.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        NULL::uuid     AS event_definition_id,
        NULL::uuid     AS bundle_id,
        NULL::uuid     AS application_id,
        vendors.id     AS vendor_id,
        NULL::uuid     AS product_id,
        expand.key     AS key,
        elements.value AS value
 FROM vendors,
      jsonb_each(vendors.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid      AS package_id,
        NULL::uuid      AS api_definition_id,
        NULL::uuid      AS event_definition_id,
        NULL::uuid      AS bundle_id,
        NULL::uuid      AS application_id,
        NULL::uuid      AS vendor_id,
        products.id     AS product_id,
        expand.key      AS key,
        elements.value  AS value
 FROM products,
      jsonb_each(products.documentation_labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements);

---

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

---

CREATE VIEW countries AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             elements.value AS value
      FROM packages,
           jsonb_array_elements_text(packages.countries) AS elements) AS package_countries
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        elements.value     AS value
 FROM api_definitions,
      jsonb_array_elements_text(api_definitions.countries) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_array_elements_text(event_api_definitions.countries) AS elements);

---

CREATE VIEW changelog_entries AS
SELECT *
FROM (SELECT id                      AS api_definition_id,
             NULL::uuid              AS event_definition_id,
             entries.version         AS version,
             entries."releaseStatus" AS release_status,
             entries.date            AS date,
             entries.description     AS description,
             entries.url             AS url
      FROM api_definitions,
           jsonb_to_recordset(api_definitions.changelog_entries) AS entries(version TEXT, "releaseStatus" TEXT,
                                                                            date TEXT,
                                                                            description TEXT, url TEXT)) AS api_entries
UNION ALL
(SELECT NULL::uuid              AS api_definition_id,
        id                      AS event_definition_id,
        entries.version         AS version,
        entries."releaseStatus" AS release_status,
        entries.date            AS date,
        entries.description     AS description,
        entries.url             AS url
 FROM event_api_definitions,
      jsonb_to_recordset(event_api_definitions.changelog_entries) AS entries(version TEXT, "releaseStatus" TEXT,
                                                                             date TEXT,
                                                                             description TEXT, url TEXT));

---

CREATE VIEW tags AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             elements.value AS value
      FROM packages,
           jsonb_array_elements_text(packages.tags) AS elements) AS package_tags
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        elements.value     AS value
 FROM api_definitions,
      jsonb_array_elements_text(api_definitions.tags) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_array_elements_text(event_api_definitions.tags) AS elements);

---

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

---

CREATE VIEW links AS
SELECT *
FROM (SELECT id                AS package_id,
             NULL::uuid        AS api_definition_id,
             NULL::uuid        AS event_definition_id,
             NULL::uuid        AS bundle_id,
             links.title       AS title,
             links.url         AS url,
             links.description AS description
      FROM packages,
           jsonb_to_recordset(packages.links) AS links(title TEXT, description TEXT, url TEXT)) AS package_links
UNION ALL
(SELECT NULL::uuid        AS package_id,
        id                AS api_definition_id,
        NULL::uuid        AS event_definition_id,
        NULL::uuid        AS bundle_id,
        links.title       AS title,
        links.url         AS url,
        links.description AS description
 FROM api_definitions,
      jsonb_to_recordset(api_definitions.links) AS links(title TEXT, description TEXT, url TEXT))
UNION ALL
(SELECT NULL::uuid        AS package_id,
        NULL::uuid        AS api_definition_id,
        id                AS event_definition_id,
        NULL::uuid        AS bundle_id,
        links.title       AS title,
        links.url         AS url,
        links.description AS description
 FROM event_api_definitions,
      jsonb_to_recordset(event_api_definitions.links) AS links(title TEXT, description TEXT, url TEXT))
UNION ALL
(SELECT NULL::uuid        AS package_id,
        NULL::uuid        AS api_definition_id,
        NULL::uuid        AS event_definition_id,
        id                AS bundle_id,
        links.title       AS title,
        links.url         AS url,
        links.description AS description
 FROM bundles,
      jsonb_to_recordset(bundles.links) AS links(title TEXT, description TEXT, url TEXT));

---

CREATE VIEW correlation_ids AS
SELECT *
FROM (SELECT applications.id            AS application_id,
             NULL::uuid                 AS product_id,
             NULL::uuid                 AS bundle_id,
             elements.value             AS value
      FROM applications,
           jsonb_array_elements_text(applications.correlation_ids) AS elements) AS app_correlation_ids
UNION ALL
(SELECT NULL::uuid         AS application_id,
        products.id        AS product_id,
        NULL::uuid         AS bundle_id,
        elements.value     AS value
 FROM products,
      jsonb_array_elements_text(products.correlation_ids) AS elements)
UNION ALL
(SELECT NULL::uuid        AS application_id,
        NULL::uuid        AS product_id,
        bundles.id        AS bundle_id,
        elements.value    AS value
 FROM bundles,
      jsonb_array_elements_text(bundles.correlation_ids) AS elements);

COMMIT;
