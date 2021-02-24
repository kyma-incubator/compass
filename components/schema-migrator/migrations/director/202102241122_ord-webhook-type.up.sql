BEGIN;

ALTER TABLE webhooks
    ALTER COLUMN type TYPE VARCHAR(255);

DROP TYPE webhook_type;

CREATE TYPE webhook_type AS ENUM (
    'CONFIGURATION_CHANGED',
    'REGISTER_APPLICATION',
    'UNREGISTER_APPLICATION',
    'OPEN_RESOURCE_DISCOVERY'
    );

ALTER TABLE webhooks
    ALTER COLUMN type TYPE webhook_type USING (type::webhook_type);

CREATE UNIQUE INDEX single_ord_webhook ON webhooks (app_id, type)
    WHERE type = 'OPEN_RESOURCE_DISCOVERY'::webhook_type;

ALTER TABLE packages
    DROP CONSTRAINT packages_vendor_fk;

DROP INDEX IF EXISTS products_tenant_id_ord_id_idx;
DROP INDEX IF EXISTS vendors_tenant_id_ord_id_idx;

ALTER TABLE packages
    ADD CONSTRAINT package_ord_id_unique UNIQUE (app_id, ord_id);
ALTER TABLE products
    ADD CONSTRAINT product_ord_id_unique UNIQUE (app_id, ord_id);
ALTER TABLE vendors
    ADD CONSTRAINT vendor_ord_id_unique UNIQUE (app_id, ord_id);
ALTER TABLE api_definitions
    ADD CONSTRAINT api_def_ord_id_unique UNIQUE (app_id, ord_id);
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_def_ord_id_unique UNIQUE (app_id, ord_id);
ALTER TABLE bundles
    ADD CONSTRAINT bundles_ord_id_unique UNIQUE (app_id, ord_id);

ALTER TABLE packages
    ADD CONSTRAINT packages_vendor_fk
        FOREIGN KEY (app_id, vendor)
            REFERENCES vendors (app_id, ord_id);

ALTER TABLE products
    ADD CONSTRAINT products_vendor_fk
        FOREIGN KEY (app_id, vendor)
            REFERENCES vendors (app_id, ord_id);

UPDATE api_definitions a
SET app_id = (SELECT b.app_id
              FROM bundles b
              WHERE id = (SELECT bundle_id FROM api_definitions WHERE id = a.id));

UPDATE event_api_definitions e
SET app_id = (SELECT b.app_id
              FROM bundles b
              WHERE id = (SELECT bundle_id FROM event_api_definitions WHERE id = e.id));

COMMIT;
