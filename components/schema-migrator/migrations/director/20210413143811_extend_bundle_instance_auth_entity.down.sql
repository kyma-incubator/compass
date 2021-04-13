BEGIN;

ALTER TABLE bundle_instance_auths
    DROP COLUMN runtime_id,
    DROP COLUMN runtime_context_id;

COMMIT;
