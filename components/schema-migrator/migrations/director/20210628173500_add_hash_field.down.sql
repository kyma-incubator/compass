ALTER TABLE api_definitions
    DROP COLUMN resource_hash;

ALTER TABLE event_api_definitions
    DROP COLUMN resource_hash;

ALTER TABLE packages
    DROP COLUMN resource_hash;
