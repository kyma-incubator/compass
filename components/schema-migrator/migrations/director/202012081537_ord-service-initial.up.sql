BEGIN;

ALTER TABLE packages
    ADD COLUMN ord_id            VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN short_description VARCHAR(255), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN version           VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN links             JSONB,
    ADD COLUMN terms_of_service  VARCHAR(512),
    ADD COLUMN licence_type      VARCHAR(256),
    ADD COLUMN licence           VARCHAR(512),
    ADD COLUMN provider          JSONB,
    ADD COLUMN tags              JSONB,
    ADD COLUMN actions           JSONB,
    ADD COLUMN extensions        JSONB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */

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
    ADD COLUMN documentation         VARCHAR(512),
    ADD COLUMN system_instance_aware BOOLEAN,
    ADD COLUMN api_protocol          api_protocol, /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN tags                  JSONB,
    ADD COLUMN api_definitions       JSONB, /* Array of URLs pointing to API specs */
    ADD COLUMN links                 JSONB,
    ADD COLUMN actions               JSONB,
    ADD COLUMN release_status        release_status, /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN changelog_entries     JSONB,
    ADD COLUMN extensions            JSONB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */

UPDATE api_definitions
SET name = name; /* Dummy update in order to apply the trigger logic above */

ALTER TABLE event_api_definitions
    ADD COLUMN ord_id                VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN short_description     VARCHAR(255), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN system_instance_aware BOOLEAN,
    ADD COLUMN changelog_entries     JSONB,
    ADD COLUMN links                 JSONB,
    ADD COLUMN tags                  JSONB,
    ADD COLUMN release_status        release_status, /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN event_definitions     JSONB, /* Array of URLs pointing to Event specs */
    ADD COLUMN extensions            JSONB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */

UPDATE event_api_definitions
SET name = name; /* Dummy update in order to apply the trigger logic above */

/* Views exposing JSON columns structured */

CREATE VIEW links AS
SELECT *
FROM (SELECT id                AS package_id,
             NULL::uuid        AS api_definition_id,
             NULL::uuid        AS event_definition_id,
             links.title       AS title,
             links.description AS description,
             links.url         AS url,
             links.extensions  AS extensions
      FROM packages,
           jsonb_to_recordset(packages.links) AS links(title TEXT, description TEXT, url TEXT, extensions JSONB)) AS package_links
UNION ALL
(SELECT NULL::uuid        AS package_id,
        id                AS api_definition_id,
        NULL::uuid        AS event_definition_id,
        links.title       AS title,
        links.description AS description,
        links.url         AS url,
        links.extensions  AS extensions
 FROM api_definitions,
      jsonb_to_recordset(api_definitions.links) AS links(title TEXT, description TEXT, url TEXT, extensions JSONB))
UNION ALL
(SELECT NULL::uuid        AS package_id,
        NULL::uuid        AS api_definition_id,
        id                AS event_definition_id,
        links.title       AS title,
        links.description AS description,
        links.url         AS url,
        links.extensions  AS extensions
 FROM event_api_definitions,
      jsonb_to_recordset(event_api_definitions.links) AS links(title TEXT, description TEXT, url TEXT, extensions JSONB));

CREATE VIEW providers AS
SELECT packages.id                        AS package_id,
       packages.provider ->> 'name'       AS name,
       packages.provider ->> 'department' AS department,
       packages.links ->> 'extensions'    AS extensions
FROM packages;

CREATE VIEW tags AS
SELECT *
FROM (SELECT packages.id    AS package_id,
             NULL::uuid     AS api_definition_id,
             NULL::uuid     AS event_definition_id,
             expand.key     AS key,
             elements.value AS value
      FROM packages,
           jsonb_each(packages.tags) AS expand,
           jsonb_array_elements_text(expand.value) AS elements) AS package_tags
UNION ALL
(SELECT NULL::uuid         AS package_id,
        api_definitions.id AS api_definition_id,
        NULL::uuid         AS event_definition_id,
        expand.key         AS key,
        elements.value     AS value
 FROM api_definitions,
      jsonb_each(api_definitions.tags) AS expand,
      jsonb_array_elements_text(expand.value) AS elements)
UNION ALL
(SELECT NULL::uuid     AS package_id,
        NULL::uuid     AS api_definition_id,
        id             AS event_definition_id,
        expand.key     AS key,
        elements.value AS value
 FROM event_api_definitions,
      jsonb_each(event_api_definitions.tags) AS expand,
      jsonb_array_elements_text(expand.value) AS elements);

CREATE VIEW package_actions AS
SELECT id                   AS package_id,
       actions.target       AS target,
       actions.type         AS type,
       actions."customType" AS custom_type,
       actions.description  AS description,
       actions.extensions   AS extensions
FROM packages,
     jsonb_to_recordset(packages.actions) AS actions(target TEXT, type TEXT, "customType" TEXT, description TEXT,
                                                     extensions JSONB);

CREATE VIEW api_resource_definitions AS
SELECT id                        AS api_definition_id,
       api_res_defs.type         AS type,
       api_res_defs."customType" AS custom_type,
       api_res_defs.url          AS url,
       api_res_defs."mediaType"  AS media_type,
       api_res_defs.extensions   AS extensions
FROM api_definitions,
     jsonb_to_recordset(api_definitions.api_definitions) AS api_res_defs(type TEXT, "customType" TEXT, "mediaType" TEXT,
                                                                         url TEXT, extensions JSONB);

CREATE VIEW api_actions AS
SELECT id                   AS api_definition_id,
       actions.target       AS target,
       actions.description  AS description,
       actions.type         AS type,
       actions."customType" AS custom_type,
       actions.extensions   AS extensions
FROM api_definitions,
     jsonb_to_recordset(api_definitions.actions) AS actions(target TEXT, description TEXT, type TEXT, "customType" TEXT,
                                                            extensions JSONB);

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

CREATE VIEW event_resource_definitions AS
SELECT id                          AS event_definition_id,
       event_res_defs.type         AS type,
       event_res_defs."customType" AS custom_type,
       event_res_defs.url          AS url,
       event_res_defs."mediaType"  AS media_type,
       event_res_defs.extensions   AS extensions
FROM event_api_definitions,
     jsonb_to_recordset(event_api_definitions.event_definitions) AS event_res_defs(type TEXT, "customType" TEXT,
                                                                                   "mediaType" TEXT,
                                                                                   url TEXT, extensions JSONB);

COMMIT;
