ALTER TABLE api_definitions
    ADD COLUMN resource_hash VARCHAR(255);

ALTER TABLE event_api_definitions
    ADD COLUMN resource_hash VARCHAR(255);

ALTER TABLE packages
    ADD COLUMN resource_hash VARCHAR(255);
