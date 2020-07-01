BEGIN;

ALTER TABLE gardener_config ALTER COLUMN auto_update_kubernetes_version DROP NOT NULL;

ALTER TABLE gardener_config ALTER COLUMN auto_update_machine_image_version DROP NOT NULL;

COMMIT;
