BEGIN;

ALTER TABLE webhooks
    ALTER COLUMN type TYPE VARCHAR(255);

DROP TYPE webhook_type;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'DELETE_APPLICATION',
    'OPEN_RESOURCE_DISCOVERY'
    );

ALTER TABLE webhooks
    ALTER COLUMN type TYPE webhook_type USING (type::webhook_type);


ALTER TABLE packages
    ADD CONSTRAINT ord_id_unique UNIQUE (ord_id);
ALTER TABLE api_definitions
    ADD CONSTRAINT ord_id_unique UNIQUE (ord_id);
ALTER TABLE event_api_definitions
    ADD CONSTRAINT ord_id_unique UNIQUE (ord_id);
ALTER TABLE bundles
    ADD CONSTRAINT ord_id_unique UNIQUE (ord_id);

COMMIT;
