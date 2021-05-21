BEGIN;

ALTER TABLE bundle_instance_auths ADD COLUMN runtime_id uuid REFERENCES runtimes(id) ON DELETE SET NULL;
ALTER TABLE bundle_instance_auths ADD COLUMN runtime_context_id uuid REFERENCES runtime_contexts(id) ON DELETE SET NULL;

COMMIT;
