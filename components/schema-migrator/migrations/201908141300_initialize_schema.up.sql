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

-- Webhook

CREATE TYPE webhooks_type AS ENUM (
    'CONFIGURATION_CHANGED'
);

CREATE TABLE webhooks (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid REFERENCES applications (id) ON DELETE CASCADE NOT NULL,
    url varchar(256) NOT NULL,
    type webhooks_type NOT NULL,
    auth jsonb
);

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
    app_id uuid REFERENCES applications (id) ON DELETE CASCADE NOT NULL,
    name varchar(256) NOT NULL,
    description text,
    group varchar(256),
    spec_data text,
    spec_format api_spec_format,
    spec_type api_spec_type,
    target_url varchar(256) NOT NULL,
    default_auth jsonb,
    version_value varchar(256),
    version_deprecated bool,
    version_deprecated_since varchar(256),
    version_for_removal bool
);

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
    app_id uuid REFERENCES applications (id) ON DELETE CASCADE NOT NULL,
    name varchar(256) NOT NULL,
    description text,
    group varchar(256),
    spec_data text,
    spec_format event_api_spec_format,
    spec_type event_api_spec_type,
    version_value varchar(256),
    version_deprecated bool,
    version_deprecated_since varchar(256),
    version_for_removal bool
);

-- Runtime Auth

CREATE TABLE runtime_auths (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    runtime_id uuid REFERENCES runtimes (id) ON DELETE CASCADE NOT NULL,
    app_def_id uuid REFERENCES api_definitions (id) ON DELETE CASCADE NOT NULL,
    value jsonb
);

-- Document

CREATE TYPE document_format AS ENUM (
    'MARKDOWN'
);

CREATE TABLE documents (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    app_id uuid REFERENCES applications (id) ON DELETE CASCADE NOT NULL,
    title varchar(256),
    display_name varchar(256),
    description text,
    format document_format,
    kind varchar(256),
    data text
);

-- Label Definition

CREATE TABLE label_definitions (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    schema jsonb
);

CREATE UNIQUE INDEX ON label_definitions (tenant_id, key);

-- Label

CREATE TABLE labels (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    app_id uuid REFERENCES applications (id) ON DELETE CASCADE,
    runtime_id uuid REFERENCES runtimes (id) ON DELETE CASCADE,
    value jsonb
);

CREATE UNIQUE INDEX ON labels (tenant_id, key, runtime_id, app_id);

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
    document_id uuid REFERENCES documents (id) ON DELETE CASCADE,
    api_def_id uuid REFERENCES api_definitions (id) ON DELETE CASCADE,
    event_api_def_id uuid REFERENCES event_api_definitions (id) ON DELETE CASCADE,
    url varchar(256) NOT NULL,
    auth jsonb,
    mode fetch_request_mode NOT NULL,
    filter varchar(256),
    status_condition fetch_request_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL
);

ALTER TABLE api_definitions
    ADD fetch_request_id uuid REFERENCES fetch_requests(id) ON DELETE CASCADE;
ALTER TABLE documents
    ADD fetch_request_id uuid REFERENCES fetch_requests(id) ON DELETE CASCADE;
ALTER TABLE event_api_definitions
    ADD fetch_request_id uuid REFERENCES fetch_requests(id) ON DELETE CASCADE;