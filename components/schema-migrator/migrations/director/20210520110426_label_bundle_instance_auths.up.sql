BEGIN;

ALTER TABLE labels ADD COLUMN bundle_instance_auth_id uuid;

DROP INDEX IF EXISTS labels_tenant_id_key_coalesce_coalesce1_coalesce2_idx;
CREATE UNIQUE INDEX labels_default_values ON labels (tenant_id, key, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_context_id, '00000000-0000-0000-0000-000000000000'), coalesce(bundle_instance_auth_id, '00000000-0000-0000-0000-000000000000'));

ALTER TABLE labels
    DROP CONSTRAINT valid_refs;

ALTER TABLE labels
    ADD CONSTRAINT valid_refs
        CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL OR runtime_context_id IS NOT NULL OR bundle_instance_auth_id IS NOT NULL);

CREATE VIEW bundle_instance_auths_with_labels AS
    SELECT labels.key, labels.value, bundles.id as bundle_id, bundles.app_id, bundle_instance_auths.runtime_id, bundle_instance_auths.status_reason as status
    FROM labels
    INNER JOIN bundle_instance_auths ON labels.bundle_instance_auth_id = bundle_instance_auths.id
    INNER JOIN bundles ON bundles.id = bundle_instance_auths.bundle_id;

COMMIT;
