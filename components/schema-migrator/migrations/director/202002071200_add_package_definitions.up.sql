CREATE TABLE packages (
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id UUID NOT NULL,
    app_id uuid NOT NULL,
    CONSTRAINT packages_tenant_id_fkey FOREIGN KEY (tenant_id, app_id) REFERENCES applications (tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT packages_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id),
    name VARCHAR(256) NOT NULL,
    description TEXT,
    instance_auth_request_json_schema JSONB,
    default_instance_auth JSONB
);

CREATE INDEX ON packages (tenant_id);
CREATE UNIQUE INDEX ON packages (tenant_id, id);

CREATE TYPE package_instance_auth_status_condition AS ENUM (
    'PENDING',
    'SUCCEEDED',
    'FAILED',
    'UNUSED'
);

CREATE TABLE package_instance_auths (
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    package_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    CONSTRAINT package_instance_auths_tenant_id_fkey FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    CONSTRAINT package_instance_auths_tenant_id_fkey1 FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id),
    context JSONB,
    input_params JSONB,
    auth_value JSONB,
    status_condition package_instance_auth_status_condition NOT NULL,
    status_timestamp TIMESTAMP NOT NULL,
    status_message VARCHAR(256) NOT NULL,
    status_reason VARCHAR(256) NOT NULL
);

CREATE INDEX ON package_instance_auths (tenant_id);
CREATE UNIQUE INDEX ON package_instance_auths (tenant_id, id);
CREATE UNIQUE INDEX ON package_instance_auths (tenant_id, package_id, id);

ALTER TABLE api_definitions ADD COLUMN package_id uuid; -- TODO: Make it NOT NULL
ALTER TABLE api_definitions ALTER COLUMN app_id DROP NOT NULL;
ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_package_id_fk
        FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id) ON DELETE CASCADE;

ALTER TABLE event_api_definitions ADD COLUMN package_id uuid; -- TODO: Make it NOT NULL
ALTER TABLE event_api_definitions ALTER COLUMN app_id DROP NOT NULL;
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_package_id_fk
        FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id) ON DELETE CASCADE;

ALTER TABLE documents ADD COLUMN package_id uuid; -- TODO: Make it NOT NULL
ALTER TABLE documents ALTER COLUMN app_id DROP NOT NULL;
ALTER TABLE documents
    ADD CONSTRAINT documents_package_id_fk
        FOREIGN KEY (tenant_id, package_id) REFERENCES packages (tenant_id, id) ON DELETE CASCADE;
