ALTER TABLE operation DROP COLUMN stage;
ALTER TABLE operation DROP COLUMN last_transition;

DROP TABLE runtime_upgrade;
DROP TYPE runtime_upgrade_state;

ALTER TABLE kyma_config ADD CONSTRAINT kyma_config_cluster_id_key UNIQUE (cluster_id);
