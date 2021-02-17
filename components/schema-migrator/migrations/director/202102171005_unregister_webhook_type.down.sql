BEGIN;

ALTER TYPE webhook_type ADD VALUE 'DELETE_APPLICATION';

UPDATE webhooks SET type = 'DELETE_APPLICATION' WHERE type = 'UNREGISTER_APPLICATION';

ALTER TYPE webhook_type RENAME TO webhook_type_old;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'DELETE_APPLICATION'
);

ALTER TABLE webhooks ALTER COLUMN type TYPE webhook_type USING type::text::webhook_type;

DROP TYPE webhook_type_old;

COMMIT;
