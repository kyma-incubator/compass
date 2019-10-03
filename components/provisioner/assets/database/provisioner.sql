-- Cluster

CREATE TABLE cluster
(
    ID uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    kubeconfig text,
    terraform_state json,
    creation_timestamp timestamp without time zone NOT NULL
);


-- Cluster Config

CREATE TYPE infrastructure_provider AS ENUM (
    'GARDENER',
    'GCP',
    'AKS'
    );

CREATE TABLE cluster_config
(
    ID uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name varchar(256) NOT NULL,
    node_count integer NOT NULL,
    disk_size varchar(256) NOT NULL,
    machine_type varchar(256) NOT NULL,
    compute_zone varchar(256) NOT NULL,
    version varchar(256) NOT NULL,
    infrastructure_provider infrastructure_provider NOT NULL,
    runtime_id uuid NOT NULL,
    foreign key (runtime_id) REFERENCES cluster (id) ON DELETE CASCADE
);


-- Infrastructure Provider Config

CREATE TABLE infrastructure_provider_config
(
    ID uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    key varchar(256) NOT NULL,
    value varchar(256) NOT NULL,
    cluster_config_id uuid NOT NULL,
    foreign key (cluster_config_id) REFERENCES cluster_config (id) ON DELETE CASCADE
);


-- Operation

CREATE TYPE operation_state AS ENUM (
    'IN_PROGRESS',
    'SUCCEEDED',
    'FAILED'
    );

CREATE TYPE operation_type AS ENUM (
    'PROVISION',
    'UPGRADE',
    'DEPROVISION',
    'RECONNECT_RUNTIME'
    );

CREATE TABLE operation
(
    ID uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    type operation_type NOT NULL,
    state operation_state NOT NULL,
    message text,
    start_timestamp timestamp without time zone NOT NULL,
    end_timestamp timestamp without time zone NOT NULL,
    runtime_id uuid NOT NULL,
    foreign key (runtime_id) REFERENCES cluster (id) ON DELETE CASCADE
);


-- Kyma Config

CREATE TABLE kyma_config
(
    ID uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    version varchar(256) NOT NULL,
    runtime_id uuid NOT NULL,
    foreign key (runtime_id) REFERENCES cluster (id) ON DELETE CASCADE
);

CREATE TABLE kyma_config_module
(
    ID uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    module varchar(256) NOT NULL,
    kyma_config_id uuid NOT NULL,
    foreign key (kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE
);
