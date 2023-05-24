BEGIN;

DROP TABLE IF EXISTS app_template_versions;

ALTER TABLE api_definitions
    DROP CONSTRAINT api_definitions_app_template_version_id_fk;
ALTER TABLE api_definitions
    DROP COLUMN app_template_version_id;

ALTER TABLE bundles
    DROP CONSTRAINT app_template_version_id;
ALTER TABLE bundles
    DROP COLUMN app_template_version_id;

ALTER TABLE documents
    DROP CONSTRAINT documents_app_template_version_id_fk;
ALTER TABLE documents
    DROP COLUMN app_template_version_id;

ALTER TABLE event_api_definitions
    DROP CONSTRAINT app_template_version_id;
ALTER TABLE event_api_definitions
    DROP COLUMN app_template_version_id;

ALTER TABLE packages
    DROP CONSTRAINT packages_app_template_version_id_fk;
ALTER TABLE packages
    DROP COLUMN app_template_version_id;

ALTER TABLE products
    DROP CONSTRAINT products_app_template_version_id_fk;
ALTER TABLE products
    DROP COLUMN app_template_version_id;

ALTER TABLE tombstones
    DROP CONSTRAINT tombstones_app_template_version_id_fk;
ALTER TABLE tombstones
    DROP COLUMN app_template_version_id;

ALTER TABLE vendors
    DROP CONSTRAINT vendors_app_template_version_id_fk;
ALTER TABLE vendors
    DROP COLUMN app_template_version_id;

COMMIT;