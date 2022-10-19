BEGIN;

ALTER TYPE tenant_type RENAME TO tenant_type_old;

CREATE TYPE tenant_type AS ENUM (
    'unknown',
    'account',
    'customer'
    'subaccount'
    );

ALTER TABLE business_tenant_mappings ALTER COLUMN type TYPE tenant_type USING type::text::tenant_type;

DROP TYPE tenant_type_old;

COMMIT;

