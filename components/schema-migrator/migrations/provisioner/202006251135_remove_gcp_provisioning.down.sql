BEGIN;

ALTER TABLE cluster ADD COLUMN terraform_state bytea;
ALTER TABLE cluster ADD COLUMN credentials_secret_name varchar(256) NOT NULL default '';

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

COMMIT;
