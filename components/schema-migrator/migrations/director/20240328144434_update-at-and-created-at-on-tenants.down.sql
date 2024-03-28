BEGIN;

-- Drop the triggers
DROP TRIGGER set_created_at_on_business_tenant_mappings ON business_tenant_mappings;
DROP TRIGGER set_updated_at_on_business_tenant_mappings ON business_tenant_mappings;

-- Drop the helper funcitons
DROP FUNCTION set_created_at();
DROP FUNCTION set_updated_at();

-- Drop the columns
ALTER TABLE business_tenant_mappings
    DROP COLUMN created_at,
    DROP COLUMN updated_at;

COMMIT;
