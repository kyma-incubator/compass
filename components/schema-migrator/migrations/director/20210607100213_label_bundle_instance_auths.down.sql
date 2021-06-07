BEGIN;

DROP INDEX IF EXISTS labels_default_values;
CREATE UNIQUE INDEX ON labels (tenant_id, key, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_context_id, '00000000-0000-0000-0000-000000000000'));

ALTER TABLE labels
DROP CONSTRAINT valid_refs;

ALTER TABLE labels
    ADD CONSTRAINT valid_refs
        CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL);

DROP VIEW bundle_instance_auths_with_labels;

ALTER TABLE labels DROP COLUMN bundle_instance_auth_id;

COMMIT;
