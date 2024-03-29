BEGIN;

-- Add columns
ALTER TABLE business_tenant_mappings
    ADD COLUMN created_at timestamp DEFAULT NOW(),
    ADD COLUMN updated_at timestamp DEFAULT NOW();

-- Introduce funciton that sets the created_at to now
CREATE OR REPLACE FUNCTION set_created_at()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.created_at = NOW();
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Define a trigger that uses funciton for created_at on business_tenant_mappings table
CREATE TRIGGER set_created_at_on_business_tenant_mappings
    BEFORE INSERT 
    ON business_tenant_mappings
    FOR EACH ROW
EXECUTE FUNCTION set_created_at();

-- Introduce funciton that sets the updated_at to now
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Define a trigger that uses funciton for updated_at on business_tenant_mappings table
CREATE TRIGGER set_updated_at_on_business_tenant_mappings
    BEFORE INSERT OR UPDATE 
    ON business_tenant_mappings
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMIT;
