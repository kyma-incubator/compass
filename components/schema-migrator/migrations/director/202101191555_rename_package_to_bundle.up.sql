BEGIN;

ALTER TABLE fetch_requests
    ALTER COLUMN mode TYPE VARCHAR(255);
DROP TYPE fetch_request_mode;

CREATE TYPE fetch_request_mode AS ENUM (
    'SINGLE',
    'BUNDLE',
    'INDEX'
    );

ALTER TABLE fetch_requests
    ALTER COLUMN mode TYPE fetch_request_mode USING (mode::fetch_request_mode);

DROP VIEW links;
DROP VIEW providers;
DROP VIEW ord_labels;
DROP VIEW tags;
DROP VIEW countries;
DROP VIEW package_links;


ALTER TABLE packages
    DROP COLUMN version,
    DROP COLUMN package_links,
    DROP COLUMN licence_type,
    DROP COLUMN provider,
    DROP COLUMN tags,
    DROP COLUMN countries;

ALTER TABLE packages
    RENAME TO bundles;

ALTER TABLE bundles
    ADD COLUMN credential_request_strategies JSONB;

ALTER TABLE api_definitions
    RENAME package_id TO bundle_id;
ALTER TABLE api_definitions
    DROP CONSTRAINT api_definitions_package_id_fk;
ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE;

ALTER TABLE documents
    RENAME package_id TO bundle_id;
ALTER TABLE documents
    DROP CONSTRAINT documents_package_id_fk;
ALTER TABLE documents
    ADD CONSTRAINT documents_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE;

ALTER TABLE event_api_definitions
    RENAME package_id TO bundle_id;
ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_api_definitions_package_id_fk;
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE;

ALTER TABLE package_instance_auths
    RENAME package_id TO bundle_id;

ALTER TABLE package_instance_auths
    RENAME TO bundle_instance_auths;

CREATE TABLE packages
(
    id                uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id         uuid         NOT NULL,
    ord_id            VARCHAR(256) NOT NULL,
    title             VARCHAR(256) NOT NULL,
    short_description VARCHAR(255) NOT NULL,
    description       TEXT         NOT NULL,
    version           VARCHAR(256) NOT NULL,
    package_links     JSONB,
    links             JSONB,
    licence_type      VARCHAR(256),
    provider          JSONB,
    tags              JSONB,
    countries         JSONB,
    labels            JSONB
);

CREATE INDEX ON packages (tenant_id);
CREATE UNIQUE INDEX ON packages (tenant_id, id);

ALTER TABLE api_definitions
    ADD COLUMN package_id uuid;
ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_package_id_fk
        FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id);

ALTER TABLE event_api_definitions
    ADD COLUMN package_id uuid;
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_package_id_fk
        FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id);

CREATE VIEW credential_request_strategies AS
SELECT id                                          AS bundle_id,
       credential_request_strategies.type          AS type,
       credential_request_strategies."callbackUrl" AS callback_url
FROM bundles,
     jsonb_to_recordset(bundles.credential_request_strategies) AS credential_request_strategies(type TEXT, "callbackUrl" TEXT);

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

CREATE VIEW providers AS
SELECT packages.id                        AS package_id,
       packages.provider ->> 'name'       AS name,
       packages.provider ->> 'department' AS department
FROM packages;

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

CREATE VIEW package_links AS
SELECT id                   AS package_id,
       actions.type         AS type,
       actions."customType" AS custom_type,
       actions.url          AS url
FROM packages,
     jsonb_to_recordset(packages.package_links) AS actions(type TEXT, "customType" TEXT, url TEXT);

COMMIT;
