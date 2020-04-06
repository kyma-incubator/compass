ALTER TABLE operation ADD COLUMN stage varchar(256);
ALTER TABLE operation ADD COLUMN last_transition timestamp without time zone;

ALTER TABLE kyma_config DROP CONSTRAINT kyma_config_cluster_id_key;

CREATE TYPE runtime_upgrade_state AS ENUM (
    'IN_PROGRESS',
    'SUCCEEDED',
    'FAILED',
    'ROLLED_BACK'
    );

CREATE TABLE runtime_upgrade
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    cluster_id uuid NOT NULL,
    state runtime_upgrade_state NOT NULL,
    pre_upgrade_kyma_config_id uuid NOT NULL,
    post_upgrade_kyma_config_id uuid NOT NULL,
    foreign key (cluster_id) REFERENCES cluster (id) ON DELETE CASCADE,
    foreign key (pre_upgrade_kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE,
    foreign key (post_upgrade_kyma_config_id) REFERENCES kyma_config (id) ON DELETE CASCADE
);
