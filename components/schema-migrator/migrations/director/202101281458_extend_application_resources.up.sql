BEGIN;

ALTER TABLE applications
    ADD COLUMN ready bool NOT NULL DEFAULT TRUE,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE bundles
    ADD COLUMN ready bool NOT NULL DEFAULT TRUE,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE api_definitions
    ADD COLUMN ready bool NOT NULL DEFAULT TRUE,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE event_api_definitions
    ADD COLUMN ready bool NOT NULL DEFAULT TRUE,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE documents
    ADD COLUMN ready bool NOT NULL DEFAULT TRUE,
    ADD COLUMN created_at timestamp NOT NULL,
    ADD COLUMN updated_at timestamp NOT NULL,
    ADD COLUMN deleted_at timestamp NOT NULL,
    ADD COLUMN error jsonb;

ALTER TABLE applications
    ADD CONSTRAINT application_id_ready_unique UNIQUE (id, ready);

ALTER TABLE bundles
    ADD CONSTRAINT bundle_id_ready_unique UNIQUE (id, ready);

ALTER TABLE api_definitions
    ADD CONSTRAINT api_id_ready_unique UNIQUE (id, ready);

ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_id_ready_unique UNIQUE (id, ready);

ALTER TABLE documents
    ADD CONSTRAINT document_id_ready_unique UNIQUE (id, ready);

ALTER TABLE bundles
    ADD CONSTRAINT bundles_applications_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON UPDATE CASCADE;

ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_applications_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON UPDATE CASCADE;

ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_applications_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON UPDATE CASCADE;

ALTER TABLE documents
    ADD CONSTRAINT documents_applications_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON UPDATE CASCADE;

ALTER TYPE webhook_type RENAME TO webhook_type_old;
CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'DELETE_APPLICATION'
    );
ALTER TABLE webhooks ALTER COLUMN type TYPE webhook_type USING type::text::webhook_type;
DROP TYPE webhook_type_old;

CREATE TYPE webhook_mode AS ENUM (
    'ASYNC',
    'SYNC'
);

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
    ADD COLUMN runtime_id uuid,
    ADD COLUMN integration_system_id uuid;

ALTER TABLE webhooks ALTER COLUMN app_id DROP NOT NULL;

ALTER TABLE webhooks
    ADD CONSTRAINT webhooks_runtime_id_fkey FOREIGN KEY (tenant_id, runtime_id) REFERENCES runtimes (tenant_id, id) ON DELETE CASCADE,
    ADD CONSTRAINT webhooks_integration_system_id_fkey FOREIGN KEY (integration_system_id) REFERENCES integration_systems (id) ON DELETE CASCADE;

COMMIT;
