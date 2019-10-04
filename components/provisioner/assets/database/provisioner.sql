-- Cluster

CREATE TABLE "Cluster"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    kubeconfig text,
    terraform_state jsonb,
    creation_timestamp timestamp without time zone NOT NULL
);


-- Cluster Config

CREATE TABLE "InfrastructureProvider"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name varchar(256) NOT NULL,
    schema jsonb NOT NULL
);

CREATE TABLE "ClusterConfig"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name varchar(256) NOT NULL,
    node_count integer NOT NULL,
    disk_size varchar(256) NOT NULL,
    machine_type varchar(256) NOT NULL,
    compute_zone varchar(256) NOT NULL,
    version varchar(256) NOT NULL,
    infrastructure_provider_id uuid NOT NULL,
    foreign key (infrastructure_provider_id) REFERENCES "InfrastructureProvider" (id) ON DELETE CASCADE,
    cluster_id uuid NOT NULL,
    foreign key (cluster_id) REFERENCES "Cluster" (id) ON DELETE CASCADE
);


-- Operation

CREATE TYPE "OperationState" AS ENUM (
    'IN_PROGRESS',
    'SUCCEEDED',
    'FAILED'
    );

CREATE TYPE "OperationType" AS ENUM (
    'PROVISION',
    'UPGRADE',
    'DEPROVISION',
    'RECONNECT_RUNTIME'
    );

CREATE TABLE "Operation"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    type "OperationType" NOT NULL,
    state "OperationState" NOT NULL,
    message text,
    start_timestamp timestamp without time zone NOT NULL,
    end_timestamp timestamp without time zone NOT NULL,
    cluster_id uuid NOT NULL,
    foreign key (cluster_id) REFERENCES "Cluster" (id) ON DELETE CASCADE
);


-- Kyma Config

CREATE TABLE "KymaConfig"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    version varchar(256) NOT NULL,
    cluster_id uuid NOT NULL,
    foreign key (cluster_id) REFERENCES "Cluster" (id) ON DELETE CASCADE
);

CREATE TABLE "KymaConfigModule"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    module varchar(256) NOT NULL,
    kyma_config_id uuid NOT NULL,
    foreign key (kyma_config_id) REFERENCES "KymaConfig" (id) ON DELETE CASCADE
);
