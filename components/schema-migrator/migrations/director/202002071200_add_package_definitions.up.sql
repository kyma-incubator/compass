CREATE TABLE package_definitions (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    description text,
    auth_request_json_schema jsonb,
    default_auth jsonb
);

CREATE INDEX ON package_definitions (tenant_id);
CREATE UNIQUE INDEX ON package_definitions (tenant_id, id);

CREATE TABLE api_instance_auths (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    package_id uuid NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    FOREIGN KEY (tenant_id, package_id) REFERENCES package_definitions (tenant_id, id) ON DELETE CASCADE,
    context jsonb,
    auth_value jsonb
);

CREATE INDEX ON api_instance_auths (tenant_id);
CREATE UNIQUE INDEX ON api_instance_auths (tenant_id, id);
CREATE UNIQUE INDEX ON api_instance_auths (tenant_id, package_id, id);

CREATE TYPE api_instance_auth_status_condition AS ENUM (
    'PENDING',
    'SUCCEEDED',
    'FAILED'
);

CREATE TABLE api_instance_auth_statuses (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    api_instance_auth_id uuid CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id uuid NOT NULL,
    FOREIGN KEY (tenant_id, api_instance_auth_id) REFERENCES api_instance_auths (tenant_id, id) ON DELETE CASCADE,
    condition api_instance_auth_status_condition NOT NULL,
    timestamp timestamp NOT NULL,
    message text NOT NULL,
    reason text NOT NULL
);

CREATE INDEX ON api_instance_auth_statuses (tenant_id);
CREATE UNIQUE INDEX ON api_instance_auth_statuses (tenant_id, id);
CREATE UNIQUE INDEX ON api_instance_auth_statuses (tenant_id, api_instance_auth_id, id);

ALTER TABLE api_definitions ADD COLUMN package_definition_id uuid;
ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_package_definition_id_fk
        FOREIGN KEY (package_definition_id) REFERENCES package_definitions (id) ON DELETE RESTRICT;

ALTER TABLE event_api_definitions ADD COLUMN package_definition_id uuid;
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_package_definition_id_fk
        FOREIGN KEY (package_definition_id) REFERENCES package_definitions (id) ON DELETE RESTRICT;

ALTER TABLE documents ADD COLUMN package_definition_id uuid;
ALTER TABLE documents
    ADD CONSTRAINT documents_package_definition_id_fk
        FOREIGN KEY (package_definition_id) REFERENCES package_definitions (id) ON DELETE RESTRICT;
