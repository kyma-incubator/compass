BEGIN;

ALTER TABLE api_definitions
    DROP CONSTRAINT api_definitions_app_template_version_id_fk,
    DROP CONSTRAINT api_definitions_app_template_version_id_ord_id_unique;
ALTER TABLE api_definitions
    DROP COLUMN app_template_version_id;

ALTER TABLE bundles
    DROP CONSTRAINT bundles_app_template_version_id_ord_id_unique;
ALTER TABLE bundles
    ALTER COLUMN app_id SET NOT NULL,
    DROP COLUMN app_template_version_id;

ALTER TABLE documents
    DROP CONSTRAINT documents_app_template_version_id_fk;
ALTER TABLE documents
    DROP COLUMN app_template_version_id;

ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_api_definitions_app_template_version_id_ord_id_unique;
ALTER TABLE event_api_definitions
    DROP COLUMN app_template_version_id;

ALTER TABLE packages
    DROP CONSTRAINT packages_app_template_version_id_fk,
    DROP CONSTRAINT packages_app_template_version_id_ord_id_unique;
ALTER TABLE packages
    ALTER COLUMN app_id SET NOT NULL,
    DROP COLUMN app_template_version_id;

ALTER TABLE products
    DROP CONSTRAINT products_app_template_version_id_fk;
ALTER TABLE products
    ALTER COLUMN app_id SET NOT NULL,
    DROP COLUMN app_template_version_id;

ALTER TABLE tombstones
    DROP CONSTRAINT tombstones_app_template_version_id_fk;
ALTER TABLE tombstones
    ALTER COLUMN app_id SET NOT NULL,
    DROP COLUMN app_template_version_id;

ALTER TABLE vendors
    DROP CONSTRAINT vendors_app_template_version_id_fk;
ALTER TABLE vendors
    DROP COLUMN app_template_version_id;


DROP INDEX IF EXISTS
    api_definitions_app_template_version_id,
    bundles_app_template_version_id,
    documents_app_template_version_id,
    event_api_definitions_app_template_version_id,
    packages_app_template_version_id,
    products_app_template_version_id,
    tombstones_app_template_version_id,
    vendors_app_template_version_id;

DROP TABLE IF EXISTS app_template_versions;

COMMIT;