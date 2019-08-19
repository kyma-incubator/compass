-- Runtime

CREATE TYPE runtime_status_condition AS ENUM (
    'INITIAL',
    'READY',
    'FAILED'
);

CREATE TABLE runtimes (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    description text,
    status_condition runtime_status_condition DEFAULT 'INITIAL' ::runtime_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    auth jsonb
);

ALTER TABLE runtimes
    ADD CONSTRAINT runtime_id_name_unique UNIQUE (tenant_id, name);

CREATE INDEX ON runtimes (tenant_id);
CREATE UNIQUE INDEX ON runtimes (tenant_id, id);

-- Application

CREATE TYPE application_status_condition AS ENUM (
    'INITIAL',
    'UNKNOWN',
    'READY',
    'FAILED'
);

CREATE TABLE applications (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    name varchar(36) NOT NULL,
    description text,
    status_condition application_status_condition DEFAULT 'INITIAL' ::application_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    healthcheck_url varchar(256)
);

ALTER TABLE applications
    ADD CONSTRAINT application_id_name_unique UNIQUE (tenant_id, name);

CREATE INDEX ON applications (tenant_id);
CREATE UNIQUE INDEX ON applications (tenant_id, id);

-- Webhook

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED'
);

CREATE TABLE webhooks (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    foreign key (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE,
    url varchar(256) NOT NULL,
    type webhook_type NOT NULL,
    auth jsonb
);

CREATE INDEX ON webhooks (tenant_id);
CREATE INDEX ON webhooks (tenant_id, id);

-- API Definition

CREATE TYPE api_spec_format AS ENUM (
    'YAML',
    'JSON'
);

CREATE TYPE api_spec_type AS ENUM (
    'ODATA',
    'OPEN_API'
);

CREATE TABLE api_definitions (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    foreign key (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE,
    name varchar(256) NOT NULL,
    description text,
    group_name varchar(256),
    target_url varchar(256) NOT NULL,
    spec_data text,
    spec_format api_spec_format,
    spec_type api_spec_type,
    default_auth jsonb,
    version_value varchar(256),
    version_deprecated bool,
    version_deprecated_since varchar(256),
    version_for_removal bool
);

CREATE INDEX ON api_definitions (tenant_id);
CREATE UNIQUE INDEX ON api_definitions (tenant_id, id);

-- Event API Definition

CREATE TYPE event_api_spec_format AS ENUM (
    'YAML',
    'JSON'
);

CREATE TYPE event_api_spec_type AS ENUM (
    'ASYNC_API'
);

CREATE TABLE event_api_definitions (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    foreign key (tenant_id, app_id) references applications (tenant_id, id) ON DELETE CASCADE,
    name varchar(256) NOT NULL,
    description text,
    group_name varchar(256),
    spec_data text,
    spec_format event_api_spec_format,
    spec_type event_api_spec_type,
    version_value varchar(256),
    version_deprecated bool,
    version_deprecated_since varchar(256),
    version_for_removal bool
);

CREATE INDEX ON event_api_definitions (tenant_id);
CREATE UNIQUE INDEX ON event_api_definitions (tenant_id, id);

-- Runtime Auth

CREATE TABLE runtime_auths (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    runtime_id uuid NOT NULL,
    foreign key (tenant_id, runtime_id) references runtimes (tenant_id, id) ON DELETE CASCADE,
    app_def_id uuid NOT NULL,
    foreign key (tenant_id, app_def_id) references api_definitions (tenant_id, id) ON DELETE CASCADE,
    value jsonb
);

CREATE INDEX ON runtime_auths (tenant_id);
CREATE UNIQUE INDEX ON runtime_auths (tenant_id, runtime_id, app_def_id);
CREATE UNIQUE INDEX ON runtime_auths (tenant_id, id);

-- Document

CREATE TYPE document_format AS ENUM (
    'MARKDOWN'
);

CREATE TABLE documents (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    foreign key (tenant_id, app_id) references applications (tenant_id, id) ON DELETE CASCADE,
    title varchar(256) NOT NULL,
    display_name varchar(256) NOT NULL,
    description text NOT NULL,
    format document_format NOT NULL,
    kind varchar(256),
    data text
);

CREATE INDEX ON documents (tenant_id);
CREATE UNIQUE INDEX ON documents (tenant_id, id);

-- Label Definition

CREATE TABLE label_definitions (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    schema jsonb
);

CREATE INDEX ON label_definitions (tenant_id);
CREATE UNIQUE INDEX ON label_definitions (tenant_id, key);
CREATE UNIQUE INDEX ON label_definitions (tenant_id, id);

-- Label

CREATE TABLE labels (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid, -- TODO: Update when Applications switch to DB repository
    runtime_id uuid,
    foreign key (tenant_id, runtime_id) references runtimes (tenant_id, id) ON DELETE CASCADE,
    key varchar(256) NOT NULL,
    value jsonb,
    CONSTRAINT valid_refs CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL)
);

CREATE INDEX ON labels (tenant_id);
CREATE UNIQUE INDEX ON labels (tenant_id, key, runtime_id, app_id);
CREATE UNIQUE INDEX ON labels (tenant_id, id);

-- Fetch Request

CREATE TYPE fetch_request_status_condition AS ENUM (
    'INITIAL',
    'SUCCEEDED',
    'FAILED'
);

CREATE TYPE fetch_request_mode AS ENUM (
    'SINGLE',
    'PACKAGE',
    'INDEX'
);

CREATE TABLE fetch_requests (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,

    api_def_id uuid,
    foreign key (tenant_id, api_def_id) references api_definitions (tenant_id, id) ON DELETE CASCADE,
    event_api_def_id uuid,
    foreign key (tenant_id, event_api_def_id) references event_api_definitions (tenant_id, id) ON DELETE CASCADE,
    document_id uuid,
    foreign key (tenant_id, document_id) references documents (tenant_id, id) ON DELETE CASCADE,

    url varchar(256) NOT NULL,
    auth jsonb,
    mode fetch_request_mode NOT NULL,
    filter varchar(256),
    status_condition fetch_request_status_condition DEFAULT 'INITIAL' ::fetch_request_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    CONSTRAINT valid_refs CHECK (api_def_id IS NOT NULL OR event_api_def_id IS NOT NULL OR document_id IS NOT NULL)
);

CREATE INDEX ON fetch_requests (tenant_id);
CREATE UNIQUE INDEX ON fetch_requests (tenant_id, api_def_id, event_api_def_id, document_id);
CREATE UNIQUE INDEX ON fetch_requests (tenant_id, id);

ALTER TABLE api_definitions
    ADD fetch_request_id uuid,
    ADD foreign key (tenant_id, fetch_request_id) references fetch_requests (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE documents
    ADD fetch_request_id uuid,
    ADD foreign key (tenant_id, fetch_request_id) references fetch_requests (tenant_id, id) ON DELETE CASCADE;
ALTER TABLE event_api_definitions
    ADD fetch_request_id uuid,
    ADD foreign key (tenant_id, fetch_request_id) references fetch_requests (tenant_id, id) ON DELETE CASCADE;