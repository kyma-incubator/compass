-- Runtime

CREATE TYPE runtime_status_condition AS ENUM (
    'INITIAL',
    'READY',
    'FAILED'
);

CREATE TABLE runtimes (
    id uuid NOT NULL CONSTRAINT runtime_pk PRIMARY KEY,
    tenant_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    description text,
    status_condition runtime_status_condition DEFAULT 'INITIAL' ::runtime_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    auth jsonb
);

ALTER TABLE runtimes
    ADD CONSTRAINT runtimes_id_name_unique UNIQUE (tenant_id, name);


-- Application

CREATE TYPE application_status_condition AS ENUM (
    'INITIAL',
    'UNKNOWN',
    'READY',
    'FAILED'
    );

CREATE TABLE applications (
                          id uuid NOT NULL CONSTRAINT application_pk PRIMARY KEY,
                          tenant_id uuid NOT NULL,
                          name varchar(36) NOT NULL,
                          description text,
                          status_condition application_status_condition DEFAULT 'INITIAL' ::application_status_condition NOT NULL,
                          status_timestamp timestamp NOT NULL,
                          healthcheck_url varchar(256)
);


-- Webhooks

-- TODO:

-- API Definitions

-- TODO:

-- Event API Definitions

-- TODO:

-- Documents

-- TODO:

-- Fetch Requests

-- TODO:

-- Runtime Auths

-- TODO:


-- Label Definition

CREATE TABLE label_definitions (
                                   id uuid PRIMARY KEY,
                                   tenant_id uuid NOT NULL,
                                   KEY varchar(256) NOT NULL,
                                   SCHEMA jsonb
);

CREATE UNIQUE INDEX ON label_definitions (tenant_id, KEY);

-- Label

CREATE TABLE labels (
                        id uuid PRIMARY KEY,
                        tenant_id uuid NOT NULL,
                        KEY varchar(256) NOT NULL,
                        app_id uuid REFERENCES applications (id) ON DELETE CASCADE,
                        runtime_id uuid REFERENCES runtimes (id) ON DELETE CASCADE,
                        value jsonb
);

CREATE UNIQUE INDEX ON labels (tenant_id, KEY, runtime_id, app_id);

