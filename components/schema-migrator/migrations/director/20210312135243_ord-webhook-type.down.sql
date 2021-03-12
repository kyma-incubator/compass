BEGIN;

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

ALTER TABLE vendors
    DROP CONSTRAINT vendors_application_tenant_fk,
    ADD CONSTRAINT vendors_tenant_id_fkey
        FOREIGN KEY (tenant_id, app_id)
            REFERENCES applications(tenant_id, id);

ALTER TABLE products
    DROP CONSTRAINT products_application_tenant_fk,
    ADD CONSTRAINT products_tenant_id_fkey
        FOREIGN KEY (tenant_id, app_id)
            REFERENCES applications(tenant_id, id);

ALTER TABLE tombstones
    DROP CONSTRAINT tombstones_application_tenant_fk,
    ADD CONSTRAINT tombstones_tenant_id_fkey
        FOREIGN KEY (tenant_id, app_id)
            REFERENCES applications(tenant_id, id);

ALTER TABLE packages
    DROP CONSTRAINT packages_application_tenant_fk,
    ADD CONSTRAINT packages_apps_fk
        FOREIGN KEY (tenant_id, app_id)
            REFERENCES applications(tenant_id, id);

ALTER TABLE api_definitions
    DROP CONSTRAINT api_definitions_tenant_bundle_id_fk;
ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE ;

ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_api_definitions_tenant_bundle_id_fk;
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_bundle_id_fk
        FOREIGN KEY (tenant_id, bundle_id) REFERENCES bundles (tenant_id, id) ON DELETE CASCADE ;

COMMIT;
