-- Runtime

CREATE TYPE runtime_status_condition AS ENUM (
    'INITIAL',
    'READY',
    'FAILED'
);

CREATE TABLE runtimes (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
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
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    name varchar(36) NOT NULL,
    description text,
    status_condition application_status_condition DEFAULT 'INITIAL' ::application_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    healthcheck_url varchar(256)
);

ALTER TABLE applications
    ADD CONSTRAINT application_tenant_id_name_unique UNIQUE (tenant_id, name);

CREATE INDEX ON applications (tenant_id);
CREATE UNIQUE INDEX ON applications (tenant_id, id);

-- Webhook

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED'
);

CREATE TABLE webhooks (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    constraint webhooks_tenant_id_fkey foreign key (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE,
    url varchar(256) NOT NULL,
    type webhook_type NOT NULL,
    auth jsonb
);

CREATE INDEX ON webhooks (tenant_id);
CREATE INDEX ON webhooks (tenant_id, id);

-- API Definition

CREATE TYPE api_spec_format AS ENUM (
    'YAML',
    'XML',
    'JSON'
);

CREATE TYPE api_spec_type AS ENUM (
    'ODATA',
    'OPEN_API'
);

CREATE TABLE api_definitions (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    constraint api_definitions_tenant_id_fkey foreign key (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE,
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

-- Event Definition

CREATE TYPE event_api_spec_format AS ENUM (
    'YAML',
    'XML',
    'JSON'
);

CREATE TYPE event_api_spec_type AS ENUM (
    'ASYNC_API'
);

CREATE TABLE event_api_definitions (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    constraint event_api_definitions_tenant_id_fkey foreign key (tenant_id, app_id) references applications (tenant_id, id) ON DELETE CASCADE,
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
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    runtime_id uuid NOT NULL,
    constraint runtime_auths_tenant_id_fkey foreign key (tenant_id, runtime_id) references runtimes (tenant_id, id) ON DELETE CASCADE,
    api_def_id uuid NOT NULL,
    constraint runtime_auths_tenant_id_fkey1 foreign key (tenant_id, api_def_id) references api_definitions (tenant_id, id) ON DELETE CASCADE,
    value jsonb
);

CREATE INDEX ON runtime_auths (tenant_id);
CREATE UNIQUE INDEX ON runtime_auths (tenant_id, runtime_id, api_def_id);
CREATE UNIQUE INDEX ON runtime_auths (tenant_id, id);

-- Document

CREATE TYPE document_format AS ENUM (
    'MARKDOWN'
);

CREATE TABLE documents (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    app_id uuid NOT NULL,
    constraint documents_tenant_id_fkey foreign key (tenant_id, app_id) references applications (tenant_id, id) ON DELETE CASCADE,
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
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    key varchar(256) NOT NULL,
    schema jsonb
);

CREATE INDEX ON label_definitions (tenant_id);
CREATE UNIQUE INDEX ON label_definitions (tenant_id, key);

-- Label

CREATE TABLE labels (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    app_id uuid,
    constraint labels_tenant_id_fkey foreign key (tenant_id, app_id) references applications (tenant_id, id) ON DELETE CASCADE,
    runtime_id uuid,
    constraint labels_tenant_id_fkey1 foreign key (tenant_id, runtime_id) references runtimes (tenant_id, id) ON DELETE CASCADE,
    key varchar(256) NOT NULL,
    value jsonb,
    CONSTRAINT valid_refs CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL)
);

CREATE INDEX ON labels (tenant_id);
-- We use coalesce to handle nullable columns https://stackoverflow.com/a/8289327
CREATE UNIQUE INDEX ON labels (tenant_id, key, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'));
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
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,

    api_def_id uuid,
    constraint fetch_requests_tenant_id_fkey foreign key (tenant_id, api_def_id) references api_definitions (tenant_id, id) ON DELETE CASCADE,
    event_api_def_id uuid,
    constraint fetch_requests_tenant_id_fkey1 foreign key (tenant_id, event_api_def_id) references event_api_definitions (tenant_id, id) ON DELETE CASCADE,
    document_id uuid,
    constraint fetch_requests_tenant_id_fkey2 foreign key (tenant_id, document_id) references documents (tenant_id, id) ON DELETE CASCADE,

    url varchar(256) NOT NULL,
    auth jsonb,
    mode fetch_request_mode NOT NULL,
    filter varchar(256),
    status_condition fetch_request_status_condition DEFAULT 'INITIAL' ::fetch_request_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    CONSTRAINT valid_refs CHECK (api_def_id IS NOT NULL OR event_api_def_id IS NOT NULL OR document_id IS NOT NULL)
);

CREATE INDEX ON fetch_requests (tenant_id);
-- We use coalesce to handle nullable columns https://stackoverflow.com/a/8289327
CREATE UNIQUE INDEX ON fetch_requests (tenant_id, coalesce(api_def_id, '00000000-0000-0000-0000-000000000000'), coalesce(event_api_def_id, '00000000-0000-0000-0000-000000000000'), coalesce(document_id, '00000000-0000-0000-0000-000000000000'));
CREATE UNIQUE INDEX ON fetch_requests (tenant_id, id);
