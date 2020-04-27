DROP TABLE runtime_upgrade;
DROP TYPE runtime_upgrade_state;

ALTER TABLE cluster DROP COLUMN active_kyma_config_id;

ALTER TABLE kyma_config ADD CONSTRAINT kyma_config_cluster_id_key UNIQUE (cluster_id);

ALTER TABLE operation DROP COLUMN stage;
ALTER TABLE operation DROP COLUMN last_transition;
