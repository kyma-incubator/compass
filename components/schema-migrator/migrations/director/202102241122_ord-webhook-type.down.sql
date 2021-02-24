BEGIN;

DROP INDEX IF EXISTS single_ord_webhook;

ALTER TABLE webhooks
    ALTER COLUMN type TYPE VARCHAR(255);

DROP TYPE webhook_type;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'UNREGISTER_APPLICATION'
    );

ALTER TABLE webhooks
    ALTER COLUMN type TYPE webhook_type USING (type::webhook_type);

ALTER TABLE products
    DROP CONSTRAINT products_vendor_fk;

ALTER TABLE packages
    DROP CONSTRAINT packages_vendor_fk;

ALTER TABLE packages
    DROP CONSTRAINT package_ord_id_unique;
ALTER TABLE products
    DROP CONSTRAINT product_ord_id_unique;
ALTER TABLE vendors
    DROP CONSTRAINT vendor_ord_id_unique;
ALTER TABLE api_definitions
    DROP CONSTRAINT api_def_ord_id_unique;
ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_def_ord_id_unique;
ALTER TABLE bundles
    DROP CONSTRAINT bundles_ord_id_unique;

CREATE UNIQUE INDEX ON products (tenant_id, ord_id);
CREATE UNIQUE INDEX ON vendors (tenant_id, ord_id);

ALTER TABLE packages
    ADD CONSTRAINT packages_vendor_fk
        FOREIGN KEY (tenant_id, vendor) REFERENCES vendors (tenant_id, ord_id);

COMMIT;
