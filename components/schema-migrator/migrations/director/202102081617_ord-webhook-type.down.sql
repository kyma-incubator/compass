BEGIN;

ALTER TABLE webhooks
    ALTER COLUMN type TYPE VARCHAR(255);

DROP TYPE webhook_type;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'DELETE_APPLICATION'
    );

ALTER TABLE webhooks
    ALTER COLUMN type TYPE webhook_type USING (type::webhook_type);

ALTER TABLE packages
    DROP CONSTRAINT package_ord_id_unique;
ALTER TABLE api_definitions
    DROP CONSTRAINT api_def_ord_id_unique;
ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_def_ord_id_unique;
ALTER TABLE bundles
    DROP CONSTRAINT bundles_ord_id_unique;

COMMIT;
