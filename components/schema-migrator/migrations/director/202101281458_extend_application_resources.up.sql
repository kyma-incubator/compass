BEGIN;

ALTER TABLE applications
    ADD COLUMN ready bool NOT NULL,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE bundles
    ADD COLUMN ready bool NOT NULL,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE api_definitions
    ADD COLUMN ready bool NOT NULL,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE event_api_definitions
    ADD COLUMN ready bool NOT NULL,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE documents
    ADD COLUMN ready bool NOT NULL,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

CREATE TYPE webhook_mode AS ENUM (
    'ASYNC',
    'SYNC'
);

ALTER TYPE webhook_type RENAME TO webhook_type_old;
CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'DELETE_APPLICATION'
    );
ALTER TABLE webhooks ALTER COLUMN type TYPE webhook_type USING type::text::webhook_type;
DROP TYPE webhook_type_old;

ALTER TABLE webhooks
    ADD COLUMN mode webhook_mode NOT NULL,
    ADD COLUMN correlation_id_key varchar(256) NOT NULL,
    ADD COLUMN retry_interval int,
    ADD COLUMN timeout int,
    ADD COLUMN url_template jsonb,
    ADD COLUMN input_template jsonb,
    ADD COLUMN header_template jsonb,
    ADD COLUMN output_template jsonb,
    ADD COLUMN status_template jsonb,
    ADD COLUMN runtime_id uuid NOT NULL,
    ADD COLUMN integration_system_id uuid NOT NULL;

ALTER TABLE webhooks
    ADD CONSTRAINT webhooks_runtime_id_fkey FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id) ON DELETE CASCADE,
    ADD CONSTRAINT webhooks_integration_system_id_fkey FOREIGN KEY (integration_system_id) REFERENCES integration_systems (id) ON DELETE CASCADE;

COMMIT;
