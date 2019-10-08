-- Cluster

CREATE TABLE "Cluster"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    kubeconfig text,
    terraform_state jsonb,
    creation_timestamp timestamp without time zone NOT NULL
);


-- Cluster Config

CREATE TABLE "GardenerConfig"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    cluster_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    kubernetesVersion varchar(256) NOT NULL,
    nodeCount integer NOT NULL,
    volumeSize varchar(256) NOT NULL,
    machineType varchar(256) NOT NULL,
    region varchar(256) NOT NULL,
    targetProvider varchar(256) NOT NULL,
    targetSecret varchar(256) NOT NULL,
    diskType varchar(256) NOT NULL,
    zone varchar(256) NOT NULL,
    cidr varchar(256) NOT NULL,
    autoScalerMin integer NOT NULL,
    autoScalerMax integer NOT NULL,
    maxSurge integer NOT NULL,
    maxUnavailable integer NOT NULL,
    UNIQUE(cluster_id),
    foreign key (cluster_id) REFERENCES "Cluster" (id) ON DELETE CASCADE
);

CREATE TABLE "GCPConfig"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    cluster_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    kubernetesVersion varchar(256) NOT NULL,
    numberOfNodes integer NOT NULL,
    bootDiskSize varchar(256) NOT NULL,
    machineType varchar(256) NOT NULL,
    region varchar(256) NOT NULL,
    zone varchar(256) NOT NULL,
    UNIQUE(cluster_id),
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
    UNIQUE(cluster_id),
    foreign key (cluster_id) REFERENCES "Cluster" (id) ON DELETE CASCADE
);

CREATE TABLE "KymaConfigModule"
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    module varchar(256) NOT NULL,
    kyma_config_id uuid NOT NULL,
    foreign key (kyma_config_id) REFERENCES "KymaConfig" (id) ON DELETE CASCADE
);
