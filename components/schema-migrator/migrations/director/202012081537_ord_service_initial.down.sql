BEGIN;

DROP VIEW links;
DROP VIEW providers;
DROP VIEW ord_labels;
DROP VIEW tags;
DROP VIEW countries;
DROP VIEW package_links;
DROP VIEW api_resource_definitions;
DROP VIEW api_resource_links;
DROP VIEW changelog_entries;
DROP VIEW event_resource_definitions;

ALTER TABLE packages
    DROP COLUMN ord_id,
    DROP COLUMN short_description,
    DROP COLUMN version,
    DROP COLUMN package_links,
    DROP COLUMN links,
    DROP COLUMN licence_type,
    DROP COLUMN provider,
    DROP COLUMN tags,
    DROP COLUMN countries,
    DROP COLUMN labels;

DROP TRIGGER set_release_status_api_def ON api_definitions;
DROP TRIGGER set_release_status_event_def ON event_api_definitions;
DROP FUNCTION set_release_status;

ALTER TABLE api_definitions
    DROP COLUMN ord_id,
    DROP COLUMN short_description,
    DROP COLUMN system_instance_aware,
    DROP COLUMN api_protocol,
    DROP COLUMN tags,
    DROP COLUMN countries,
    DROP COLUMN api_definitions,
    DROP COLUMN links,
    DROP COLUMN api_resource_links,
    DROP COLUMN release_status,
    DROP COLUMN sunset_date,
    DROP COLUMN successor,
    DROP COLUMN changelog_entries,
    DROP COLUMN labels;

ALTER TABLE event_api_definitions
    DROP COLUMN ord_id,
    DROP COLUMN short_description,
    DROP COLUMN system_instance_aware,
    DROP COLUMN changelog_entries,
    DROP COLUMN links,
    DROP COLUMN tags,
    DROP COLUMN countries,
    DROP COLUMN release_status,
    DROP COLUMN sunset_date,
    DROP COLUMN successor,
    DROP COLUMN event_definitions,
    DROP COLUMN labels;

DROP TYPE api_protocol;
DROP TYPE release_status;

COMMIT;
