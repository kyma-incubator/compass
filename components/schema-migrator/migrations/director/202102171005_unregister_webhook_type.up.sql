BEGIN;

ALTER TYPE webhook_type RENAME TO webhook_type_old;

CREATE TYPE webhook_type_med AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'UNREGISTER_APPLICATION', 
    'DELETE_APPLICATION'
);

ALTER TABLE webhooks ALTER COLUMN type TYPE webhook_type_med USING type::text::webhook_type_med;

UPDATE webhooks SET type = 'UNREGISTER_APPLICATION' WHERE type = 'DELETE_APPLICATION';

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'UNREGISTER_APPLICATION'
);

ALTER TABLE webhooks ALTER COLUMN type TYPE webhook_type USING type::text::webhook_type;

DROP TYPE webhook_type_old;
DROP TYPE webhook_type_med;

COMMIT;
