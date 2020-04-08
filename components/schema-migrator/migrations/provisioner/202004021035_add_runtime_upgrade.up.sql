ALTER TABLE operation ADD COLUMN stage varchar(256);
ALTER TABLE operation ADD COLUMN last_transition timestamp without time zone;

ALTER TABLE kyma_config DROP CONSTRAINT kyma_config_cluster_id_key;

BEGIN TRANSACTION;

ALTER TABLE cluster ADD COLUMN active_kyma_config_id uuid;
ALTER TABLE cluster ADD CONSTRAINT cluster_active_kyma_config_id_fkey foreign key (active_kyma_config_id) REFERENCES kyma_config (id) DEFERRABLE INITIALLY DEFERRED;

UPDATE cluster
SET active_kyma_config_id=subquery.id
FROM (SELECT id, cluster_id
      FROM  kyma_config) AS subquery
WHERE cluster.id=subquery.cluster_id;

END TRANSACTION;

ALTER TABLE cluster ALTER COLUMN active_kyma_config_id SET NOT NULL;

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
