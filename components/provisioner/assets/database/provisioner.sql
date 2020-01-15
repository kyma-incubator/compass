-- Cluster

CREATE TABLE cluster
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    kubeconfig text,
    terraform_state bytea,
    tenant varchar(256) NOT NULL,
    credentials_secret_name varchar(256) NOT NULL,
    creation_timestamp timestamp without time zone NOT NULL
);


-- Cluster Config

CREATE TABLE gardener_config
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    cluster_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    project_name varchar(256) NOT NULL,
    kubernetes_version varchar(256) NOT NULL,
    node_Count integer NOT NULL,
    volume_size_gb varchar(256) NOT NULL,
    machine_type varchar(256) NOT NULL,
    region varchar(256) NOT NULL,
    provider varchar(256) NOT NULL,
    seed varchar(256) NOT NULL,
    target_secret varchar(256) NOT NULL,
    disk_type varchar(256) NOT NULL,
    worker_cidr varchar(256) NOT NULL,
    auto_scaler_min integer NOT NULL,
    auto_scaler_max integer NOT NULL,
    max_surge integer NOT NULL,
    max_unavailable integer NOT NULL,
    provider_specific_config jsonb,
    UNIQUE(cluster_id),
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE
);

CREATE TABLE gcp_config
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    cluster_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    project_name varchar(256) NOT NULL,
    kubernetes_version varchar(256) NOT NULL,
    number_of_nodes integer NOT NULL,
    boot_disk_size_gb varchar(256) NOT NULL,
    machine_type varchar(256) NOT NULL,
    region varchar(256) NOT NULL,
    zone varchar(256) NOT NULL,
    UNIQUE(cluster_id),
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE
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
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    type operation_type NOT NULL,
    state operation_state NOT NULL,
    message text,
    start_timestamp timestamp without time zone NOT NULL,
    end_timestamp timestamp without time zone,
    cluster_id uuid NOT NULL,
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE
);

-- Kyma Release

CREATE TABLE kyma_release
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    version varchar(256) NOT NULL,
    tiller_yaml text NOT NULL,
    installer_yaml text NOT NULL,
    unique(version)
);

-- Kyma Config

CREATE TABLE kyma_config
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    release_id uuid NOT NULL,
    cluster_id uuid NOT NULL,
    global_configuration jsonb,
    UNIQUE(cluster_id),
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE,
    foreign key (release_id) REFERENCES kyma_release (id) ON DELETE RESTRICT
);

CREATE TABLE kyma_component_config
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    component varchar(256) NOT NULL,
    namespace varchar(256) NOT NULL,
    configuration jsonb,
    kyma_config_id uuid NOT NULL,
    foreign key (kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE
);
