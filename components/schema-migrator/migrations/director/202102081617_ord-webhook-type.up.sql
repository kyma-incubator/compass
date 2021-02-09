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
    ADD CONSTRAINT package_ord_id_unique UNIQUE (ord_id);
ALTER TABLE api_definitions
    ADD CONSTRAINT api_def_ord_id_unique UNIQUE (ord_id);
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_def_ord_id_unique UNIQUE (ord_id);
ALTER TABLE bundles
    ADD CONSTRAINT bundles_ord_id_unique UNIQUE (ord_id);

UPDATE api_definitions a
SET app_id = (SELECT b.app_id
              FROM bundles b
              WHERE id = (SELECT bundle_id FROM api_definitions WHERE id = a.id));

UPDATE event_api_definitions e
SET app_id = (SELECT b.app_id
              FROM bundles b
              WHERE id = (SELECT bundle_id FROM event_api_definitions WHERE id = e.id));

COMMIT;
