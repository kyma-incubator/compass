BEGIN;

UPDATE gardener_config SET auto_update_kubernetes_version=false WHERE auto_update_kubernetes_version IS NULL;
ALTER TABLE gardener_config ALTER COLUMN auto_update_kubernetes_version SET NOT NULL;

UPDATE gardener_config SET auto_update_machine_image_version=false WHERE auto_update_machine_image_version IS NULL;
ALTER TABLE gardener_config ALTER COLUMN auto_update_machine_image_version SET NOT NULL;

COMMIT;
