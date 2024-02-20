BEGIN;

DROP INDEX IF EXISTS business_tenant_mappings_type_idx;

DROP INDEX IF EXISTS X labels_key_idx;

DROP INDEX IF EXISTS X formation_assignments_source_idx;

DROP INDEX IF EXISTS  formation_assignments_target_idx;

COMMIT;
