BEGIN;

DROP VIEW ord_labels;
DROP VIEW ord_documentation_labels;
DROP VIEW line_of_businesses;
DROP VIEW countries;
DROP VIEW changelog_entries;
DROP VIEW tags;
DROP VIEW industries;
DROP VIEW links;
DROP VIEW correlation_ids;

--- breakdown 'ord_labels' view

CREATE VIEW ord_labels_packages AS
SELECT id                  AS package_id,
       expand.key          AS key,
       elements.value      AS value
FROM packages,
     jsonb_each(packages.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_labels_api_definitions AS
SELECT id                  AS api_definition_id,
       expand.key          AS key,
       elements.value      AS value
FROM api_definitions,
     jsonb_each(api_definitions.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_labels_event_definitions AS
SELECT id                  AS event_definition_id,
       expand.key          AS key,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_each(event_api_definitions.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_labels_bundles AS
SELECT id                  AS bundle_id,
       expand.key          AS key,
       elements.value      AS value
FROM bundles,
     jsonb_each(bundles.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_labels_applications AS
SELECT id                  AS application_id,
       expand.key          AS key,
       elements.value      AS value
FROM applications,
     jsonb_each(applications.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_labels_vendors AS
SELECT id                  AS vendor_id,
       expand.key          AS key,
       elements.value      AS value
FROM vendors,
     jsonb_each(vendors.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_labels_products AS
SELECT id                  AS product_id,
       expand.key          AS key,
       elements.value      AS value
FROM products,
     jsonb_each(products.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

--- breakdown 'ord_documentation_labels' view

CREATE VIEW ord_documentation_labels_packages AS
SELECT id                  AS package_id,
       expand.key          AS key,
       elements.value      AS value
FROM packages,
     jsonb_each(packages.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_api_definitions AS
SELECT id                  AS api_definition_id,
       expand.key          AS key,
       elements.value      AS value
FROM api_definitions,
     jsonb_each(api_definitions.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_event_definitions AS
SELECT id                  AS event_definition_id,
       expand.key          AS key,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_each(event_api_definitions.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_bundles AS
SELECT id                  AS bundle_id,
       expand.key          AS key,
       elements.value      AS value
FROM bundles,
     jsonb_each(bundles.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_applications AS
SELECT id                  AS application_id,
       expand.key          AS key,
       elements.value      AS value
FROM applications,
     jsonb_each(applications.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_vendors AS
SELECT id                  AS vendor_id,
       expand.key          AS key,
       elements.value      AS value
FROM vendors,
     jsonb_each(vendors.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

CREATE VIEW ord_documentation_labels_products AS
SELECT id                  AS product_id,
       expand.key          AS key,
       elements.value      AS value
FROM products,
     jsonb_each(products.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;

--- breakdown 'line_of_businesses' view

CREATE VIEW line_of_businesses_packages AS
SELECT id                  AS package_id,
       elements.value      AS value
FROM packages,
     jsonb_array_elements_text(packages.line_of_business) AS elements;

CREATE VIEW line_of_businesses_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.line_of_business) AS elements;

CREATE VIEW line_of_businesses_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.line_of_business) AS elements;

--- breakdown of 'countries' view

CREATE VIEW countries_packages AS
SELECT id                  AS package_id,
       elements.value      AS value
FROM packages,
     jsonb_array_elements_text(packages.countries) AS elements;

CREATE VIEW countries_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.countries) AS elements;

CREATE VIEW countries_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.countries) AS elements;

--- breakdown of 'changelog_entries' view

CREATE VIEW changelog_entries_api_definitions AS
SELECT id                      AS api_definition_id,
       entries.version         AS version,
       entries."releaseStatus" AS release_status,
       entries.date            AS date,
       entries.description     AS description,
       entries.url             AS url
FROM api_definitions,
     jsonb_to_recordset(api_definitions.changelog_entries) AS entries(version TEXT, "releaseStatus" TEXT, date TEXT, description TEXT, url TEXT);

CREATE VIEW changelog_entries_event_definitions AS
SELECT id                      AS event_definition_id,
       entries.version         AS version,
       entries."releaseStatus" AS release_status,
       entries.date            AS date,
       entries.description     AS description,
       entries.url             AS url
FROM event_api_definitions,
     jsonb_to_recordset(event_api_definitions.changelog_entries) AS entries(version TEXT, "releaseStatus" TEXT, date TEXT, description TEXT, url TEXT);

--- breakdown of 'tags' view

CREATE VIEW tags_packages AS
SELECT id                  AS package_id,
       elements.value      AS value
FROM packages,
     jsonb_array_elements_text(packages.tags) AS elements;

CREATE VIEW tags_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.tags) AS elements;

CREATE VIEW tags_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.tags) AS elements;

--- breakdown of 'industries' view

CREATE VIEW industries_packages AS
SELECT id                  AS package_id,
       elements.value      AS value
FROM packages,
     jsonb_array_elements_text(packages.industry) AS elements;

CREATE VIEW industries_api_definitions AS
SELECT id                  AS api_definition_id,
       elements.value      AS value
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.industry) AS elements;

CREATE VIEW industries_event_definitions AS
SELECT id                  AS event_definition_id,
       elements.value      AS value
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.industry) AS elements;

--- breakdown of 'links' view

CREATE VIEW links_packages AS
SELECT id                AS package_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM packages,
     jsonb_to_recordset(packages.links) AS links(title TEXT, description TEXT, url TEXT);

CREATE VIEW links_api_definitions AS
SELECT id                AS api_definition_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM api_definitions,
     jsonb_to_recordset(api_definitions.links) AS links(title TEXT, description TEXT, url TEXT);

CREATE VIEW links_event_definitions AS
SELECT id                AS event_definition_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM event_api_definitions,
     jsonb_to_recordset(event_api_definitions.links) AS links(title TEXT, description TEXT, url TEXT);

CREATE VIEW links_bundles AS
SELECT id                AS bundle_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM bundles,
     jsonb_to_recordset(bundles.links) AS links(title TEXT, description TEXT, url TEXT);

--- breakdown of 'correlation_ids' view

CREATE VIEW correlation_ids_applications AS
SELECT id                  AS application_id,
       elements.value      AS value
FROM applications,
     jsonb_array_elements_text(applications.correlation_ids) AS elements;

CREATE VIEW correlation_ids_products AS
SELECT id                  AS product_id,
       elements.value      AS value
FROM products,
     jsonb_array_elements_text(products.correlation_ids) AS elements;

CREATE VIEW correlation_ids_bundles AS
SELECT id                  AS bundle_id,
       elements.value      AS value
FROM bundles,
     jsonb_array_elements_text(bundles.correlation_ids) AS elements;

COMMIT;
