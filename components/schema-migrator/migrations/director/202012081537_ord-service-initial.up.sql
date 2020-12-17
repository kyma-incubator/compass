BEGIN;

ALTER TABLE packages
    ADD COLUMN ord_id            VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN short_description VARCHAR(255), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN version           VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN package_links     JSONB,
    ADD COLUMN links             JSONB,
    ADD COLUMN licence_type      VARCHAR(256),
    ADD COLUMN provider          JSONB,
    ADD COLUMN tags              JSONB,
    ADD COLUMN countries         JSONB,
    ADD COLUMN labels            JSONB;

CREATE TYPE api_protocol AS ENUM ('odata-v2', 'odata-v4','soap-inbound','soap-outbound','rest');
CREATE TYPE release_status AS ENUM ('beta', 'active', 'deprecated', 'decommissioned');

CREATE OR REPLACE FUNCTION set_release_status()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.release_status := CASE
                              WHEN NEW.release_status IS NOT NULL THEN NEW.release_status
                              WHEN NEW.version_for_removal THEN 'decommissioned'
                              WHEN NEW.version_deprecated THEN 'deprecated'
                              ELSE 'active' END;

    NEW.version_for_removal := CASE
                                   WHEN NEW.version_for_removal IS NULL AND NEW.release_status = 'decommissioned'
                                       THEN TRUE
                                   ELSE NEW.version_for_removal END;

    NEW.version_deprecated := CASE
                                  WHEN NEW.version_deprecated IS NULL AND NEW.release_status = 'deprecated' THEN TRUE
                                  ELSE NEW.version_deprecated END;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_release_status_api_def
    BEFORE INSERT OR UPDATE
    ON api_definitions
    FOR EACH ROW
EXECUTE PROCEDURE set_release_status();

CREATE TRIGGER set_release_status_event_def
    BEFORE INSERT OR UPDATE
    ON event_api_definitions
    FOR EACH ROW
EXECUTE PROCEDURE set_release_status();


ALTER TABLE api_definitions
    ADD COLUMN ord_id                VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN short_description     VARCHAR(255), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN system_instance_aware BOOLEAN,
    ADD COLUMN api_protocol          api_protocol, /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN tags                  JSONB,
    ADD COLUMN countries             JSONB,
    ADD COLUMN api_definitions       JSONB, /* Array of URLs pointing to API specs */
    ADD COLUMN links                 JSONB,
    ADD COLUMN api_resource_links    JSONB,
    ADD COLUMN release_status        release_status, /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN sunset_date           VARCHAR(256),
    ADD COLUMN successor             VARCHAR(256),
    ADD COLUMN changelog_entries     JSONB,
    ADD COLUMN labels                JSONB;

/* Dummy update in order to apply the trigger logic above */
UPDATE api_definitions
SET name = name;

ALTER TABLE event_api_definitions
    ADD COLUMN ord_id                VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN short_description     VARCHAR(255), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN system_instance_aware BOOLEAN,
    ADD COLUMN changelog_entries     JSONB,
    ADD COLUMN links                 JSONB,
    ADD COLUMN tags                  JSONB,
    ADD COLUMN countries             JSONB,
    ADD COLUMN release_status        release_status, /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN sunset_date           VARCHAR(256),
    ADD COLUMN successor             VARCHAR(256),
    ADD COLUMN event_definitions     JSONB, /* Array of URLs pointing to Event specs */
    ADD COLUMN labels                JSONB;

/* Dummy update in order to apply the trigger logic above */
UPDATE event_api_definitions
SET name = name;


/* Views exposing JSON columns structured */

CREATE VIEW links AS
SELECT *
FROM (SELECT id                AS package_id,
             NULL::uuid        AS api_definition_id,
             NULL::uuid        AS event_definition_id,
             links.title       AS title,
             links.url         AS url,
             links.description AS description
      FROM packages,
           jsonb_to_recordset(packages.links) AS links(title TEXT, description TEXT, url TEXT)) AS package_links
UNION ALL
(SELECT NULL::uuid        AS package_id,
        id                AS api_definition_id,
        NULL::uuid        AS event_definition_id,
        links.title       AS title,
        links.url         AS url,
        links.description AS description
 FROM api_definitions,
      jsonb_to_recordset(api_definitions.links) AS links(title TEXT, description TEXT, url TEXT))
UNION ALL
(SELECT NULL::uuid        AS package_id,
        NULL::uuid        AS api_definition_id,
        id                AS event_definition_id,
        links.title       AS title,
        links.url         AS url,
        links.description AS description
 FROM event_api_definitions,
      jsonb_to_recordset(event_api_definitions.links) AS links(title TEXT, description TEXT, url TEXT));

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
             expand.key     AS key,
             elements.value AS value
      FROM packages,
           jsonb_each(packages.labels) AS expand,
           jsonb_array_elements_text(expand.value) AS elements) AS package_tags
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        expand.key         AS key,
        elements.value     AS value
 FROM api_definitions,
      jsonb_each(api_definitions.labels) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        expand.key     AS key,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_each(event_api_definitions.labels) AS expand,
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
           jsonb_array_elements_text(packages.countries) AS elements) AS package_tags
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

/*TODO: Currently each API has only a single API Definition, once we implement the aggregator this view should be fully represented in a separate specifications table which will have the data as well*/
CREATE VIEW api_resource_definitions AS
SELECT *
FROM (SELECT id                                  AS api_definition_id,
             api_res_defs.type                   AS type,
             api_res_defs."customType"           AS custom_type,
             format('/api/%s/specification', id) AS url,
             api_res_defs."mediaType"            AS media_type
      FROM api_definitions,
           jsonb_to_recordset(api_definitions.api_definitions) AS api_res_defs(type TEXT, "customType" TEXT,
                                                                               "mediaType" TEXT,
                                                                               url TEXT)) as api_defs
UNION ALL
(SELECT id                                  AS api_definition_id,
        spec_type::text                     AS type,
        NULL                                AS custom_type,
        format('/api/%s/specification', id) AS url,
        spec_format::text                   AS media_type
 FROM api_definitions);

CREATE VIEW api_resource_links AS
SELECT id                   AS api_definition_id,
       actions.type         AS type,
       actions."customType" AS custom_type,
       actions.url          AS url
FROM api_definitions,
     jsonb_to_recordset(api_definitions.api_resource_links) AS actions(type TEXT, "customType" TEXT, url TEXT);

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
/*TODO: Currently each Event has only a single Event Definition, once we implement the aggregator this view should be fully represented in a separate specifications table which will have the data as well*/
CREATE VIEW event_resource_definitions AS
SELECT *
FROM (SELECT id                                    AS event_definition_id,
             event_res_defs.type                   AS type,
             event_res_defs."customType"           AS custom_type,
             format('/event/%s/specification', id) AS url,
             event_res_defs."mediaType"            AS media_type
      FROM event_api_definitions,
           jsonb_to_recordset(event_api_definitions.event_definitions) AS event_res_defs(type TEXT, "customType" TEXT,
                                                                                         "mediaType" TEXT,
                                                                                         url TEXT)) as event_defs
UNION ALL
(SELECT id                                    AS event_definition_id,
        spec_type::text                       AS type,
        NULL                                  AS custom_type,
        format('/event/%s/specification', id) AS url,
        spec_format::text                     AS media_type
 FROM event_api_definitions);

COMMIT;
