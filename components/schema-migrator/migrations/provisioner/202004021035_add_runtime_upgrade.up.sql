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

UPDATE operation
SET stage='StartingInstallation'
WHERE (operation.state='IN_PROGRESS' OR operation.state='FAILED') AND operation.type='PROVISION';

UPDATE operation
SET stage='Deprovisioning'
WHERE (operation.state='IN_PROGRESS' OR operation.state='FAILED') AND operation.type='DEPROVISION';

UPDATE operation
SET stage='Finished'
WHERE operation.state='SUCCEEDED';

END TRANSACTION;

ALTER TABLE cluster ALTER COLUMN active_kyma_config_id SET NOT NULL;
ALTER TABLE operation ALTER COLUMN stage SET NOT NULL;

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
