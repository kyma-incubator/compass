BEGIN;

CREATE INDEX IF NOT EXISTS business_tenant_mappings_type_idx ON business_tenant_mappings (type);

CREATE INDEX IF NOT EXISTS labels_key_idx ON labels (key);

CREATE INDEX IF NOT EXISTS formation_assignments_source_idx ON formation_assignments (source);

CREATE INDEX IF NOT EXISTS formation_assignments_target_idx ON formation_assignments (target);

ALTER TABLE business_tenant_mappings DROP CONSTRAINT IF EXISTS external_tenant_not_null_not_empty;
ALTER TABLE business_tenant_mappings
    ADD CONSTRAINT external_tenant_not_null_not_empty
        CHECK (external_tenant IS NOT NULL AND external_tenant != '');

COMMIT;
