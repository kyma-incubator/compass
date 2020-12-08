BEGIN;

ALTER TABLE packages
    DROP COLUMN ord_id,
    DROP COLUMN short_description,
    DROP COLUMN version,
    DROP COLUMN links,
    DROP COLUMN terms_of_service,
    DROP COLUMN licence_type,
    DROP COLUMN licence,
    DROP COLUMN provider,
    DROP COLUMN tags,
    DROP COLUMN actions,
    DROP COLUMN extensions;

DROP TRIGGER set_release_status_api_def ON api_definitions;
DROP TRIGGER set_release_status_event_def ON event_api_definitions;
DROP FUNCTION set_release_status;

ALTER TABLE api_definitions
    DROP COLUMN ord_id,
    DROP COLUMN short_description,
    DROP COLUMN documentation,
    DROP COLUMN system_instance_aware,
    DROP COLUMN api_protocol,
    DROP COLUMN tags,
    DROP COLUMN api_definitions,
    DROP COLUMN links,
    DROP COLUMN actions,
    DROP COLUMN release_status,
    DROP COLUMN changelog_entries,
    DROP COLUMN extensions;

ALTER TABLE event_api_definitions
    DROP COLUMN ord_id,
    DROP COLUMN short_description,
    DROP COLUMN system_instance_aware,
    DROP COLUMN changelog_entries,
    DROP COLUMN links,
    DROP COLUMN tags,
    DROP COLUMN release_status,
    DROP COLUMN event_definitions,
    DROP COLUMN extensions;

DROP TYPE api_protocol;
DROP TYPE release_status;

COMMIT;
