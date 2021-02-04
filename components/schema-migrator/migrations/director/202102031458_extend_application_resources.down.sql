BEGIN;

ALTER TABLE api_definitions
    DROP CONSTRAINT api_id_ready_unique,
    DROP CONSTRAINT api_definitions_bundles_ready_fk;

ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_api_id_ready_unique,
    DROP CONSTRAINT event_api_definitions_bundles_ready_fk;

ALTER TABLE documents
    DROP CONSTRAINT document_id_ready_unique,
    DROP CONSTRAINT documents_bundles_ready_fk;

ALTER TABLE bundles
    DROP CONSTRAINT bundle_id_ready_unique,
    DROP CONSTRAINT bundles_applications_ready_fk;

ALTER TABLE applications
    DROP COLUMN ready CASCADE,
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at,
    DROP COLUMN error;

ALTER TABLE bundles
    DROP COLUMN ready,
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at,
    DROP COLUMN error;

ALTER TABLE api_definitions
    DROP COLUMN ready,
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at,
    DROP COLUMN error;

ALTER TABLE event_api_definitions
    DROP COLUMN ready,
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at,
    DROP COLUMN error;

ALTER TABLE documents
    DROP COLUMN ready,
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at,
    DROP COLUMN error;

ALTER TYPE webhook_type RENAME TO webhook_type_old;
CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED'
);
ALTER TABLE webhooks ALTER COLUMN type TYPE webhook_type USING type::text::webhook_type;
DROP TYPE webhook_type_old;

ALTER TABLE webhooks
    DROP COLUMN mode,
    DROP COLUMN correlation_id_key,
    DROP COLUMN retry_interval,
    DROP COLUMN timeout,
    DROP COLUMN url_template,
    DROP COLUMN input_template,
    DROP COLUMN header_template,
    DROP COLUMN output_template,
    DROP COLUMN status_template,
    DROP COLUMN runtime_id,
    DROP COLUMN integration_system_id;

ALTER TABLE webhooks ALTER COLUMN app_id SET NOT NULL;

DROP TYPE webhook_mode;

COMMIT;
