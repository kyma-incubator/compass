BEGIN;

ALTER TABLE gardener_config DROP COLUMN auto_update_kubernetes_version;

ALTER TABLE gardener_config DROP COLUMN auto_update_machine_image_version;

COMMIT;
