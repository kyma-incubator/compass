-- This schema is used only in tests, after adjusting it provide new migration in schema-migrator component

-- Cluster

CREATE TABLE cluster
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    kubeconfig text,
    tenant varchar(256) NOT NULL,
    creation_timestamp timestamp without time zone NOT NULL,
    deleted boolean default false,
    sub_account_id varchar(256)
);

-- Cluster Config

CREATE TABLE gardener_config
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    cluster_id uuid NOT NULL,
    name varchar(256) NOT NULL UNIQUE,
    project_name varchar(256) NOT NULL,
    kubernetes_version varchar(256) NOT NULL,
    volume_size_gb varchar(256) NOT NULL,
    machine_type varchar(256) NOT NULL,
    region varchar(256) NOT NULL,
    provider varchar(256) NOT NULL,
    purpose varchar(256),
    licence_type varchar(256),
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
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE,
    stage varchar(256) NOT NULL,
    last_transition timestamp without time zone
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
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE,
    foreign key (release_id) REFERENCES kyma_release (id) ON DELETE RESTRICT
);

CREATE TABLE kyma_component_config
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    component varchar(256) NOT NULL,
    namespace varchar(256) NOT NULL,
    source_url varchar(256),
    configuration jsonb,
    component_order integer,
    kyma_config_id uuid NOT NULL,
    foreign key (kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE
);

-- Active Kyma Config column

ALTER TABLE cluster ADD COLUMN active_kyma_config_id uuid NOT NULL;
ALTER TABLE cluster ADD CONSTRAINT cluster_active_kyma_config_id_fkey foreign key (active_kyma_config_id) REFERENCES kyma_config (id) DEFERRABLE INITIALLY DEFERRED;


-- Runtime Upgrade

CREATE TYPE runtime_upgrade_state AS ENUM (
    'IN_PROGRESS',
    'SUCCEEDED',
    'FAILED',
    'ROLLED_BACK'
);

CREATE TABLE runtime_upgrade
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    operation_id uuid NOT NULL,
    state runtime_upgrade_state NOT NULL,
    pre_upgrade_kyma_config_id uuid NOT NULL,
    post_upgrade_kyma_config_id uuid NOT NULL,
    foreign key (operation_id) REFERENCES operation (id) ON DELETE CASCADE,
    foreign key (pre_upgrade_kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE,
    foreign key (post_upgrade_kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE
);
