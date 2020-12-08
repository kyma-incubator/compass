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
                              WHEN NEW.version_for_removal THEN 'decommissioned'
                              WHEN NEW.version_deprecated THEN 'deprecated'
                              ELSE 'active' END;

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
    ADD COLUMN release_status        release_status NOT NULL DEFAULT 'active',
    ADD COLUMN changelog_entries     JSONB,
    ADD COLUMN extensions            JSONB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */

UPDATE api_definitions SET name = name; /* Dummy update in order to apply the trigger logic above */

ALTER TABLE event_api_definitions
    ADD COLUMN ord_id                VARCHAR(256), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN short_description     VARCHAR(255), /* ORD Required, nullable due to backwards compatibility */
    ADD COLUMN system_instance_aware BOOLEAN,
    ADD COLUMN changelog_entries     JSONB,
    ADD COLUMN links                 JSONB,
    ADD COLUMN tags                  JSONB,
    ADD COLUMN release_status        release_status NOT NULL DEFAULT 'active',
    ADD COLUMN event_definitions     JSONB, /* Array of URLs pointing to Event specs */
    ADD COLUMN extensions            JSONB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */

UPDATE event_api_definitions SET name = name; /* Dummy update in order to apply the trigger logic above */

COMMIT;
