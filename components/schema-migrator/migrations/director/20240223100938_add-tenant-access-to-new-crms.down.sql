BEGIN;

DROP INDEX IF EXISTS business_tenant_mappings_type_idx;

DROP INDEX IF EXISTS labels_key_idx;

DROP INDEX IF EXISTS formation_assignments_source_idx;

DROP INDEX IF EXISTS  formation_assignments_target_idx;

ALTER TABLE business_tenant_mappings
DROP CONSTRAINT external_tenant_not_null_not_empty;

COMMIT;
