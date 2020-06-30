BEGIN;

ALTER TABLE gardener_config ADD COLUMN auto_update_kubernetes_version boolean NOT NULL DEFAULT false;
ALTER TABLE gardener_config ALTER COLUMN auto_update_kubernetes_version DROP DEFAULT;

ALTER TABLE gardener_config ADD COLUMN auto_update_machine_image_version boolean NOT NULL DEFAULT false;
ALTER TABLE gardener_config ALTER COLUMN auto_update_machine_image_version DROP DEFAULT;

COMMIT;
