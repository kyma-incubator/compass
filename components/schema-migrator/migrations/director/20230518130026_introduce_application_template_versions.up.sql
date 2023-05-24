BEGIN;

CREATE TABLE IF NOT EXISTS app_template_versions
(
    id              UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_template_id UUID NOT NULL,
    CONSTRAINT app_template_versions_app_template_id_fk FOREIGN KEY  (app_template_id) REFERENCES app_templates(id) on DELETE CASCADE,
    version         varchar(256) NOT NULL,
    title           varchar(256),
    correlation_ids JSONB,
    release_date    TIMESTAMP,
    created_at      TIMESTAMP NOT NULL
);

ALTER TABLE api_definitions
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT api_definitions_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE bundles
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT bundles_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE documents
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT documents_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE event_api_definitions
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT event_api_definitions_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE packages
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT packages_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE products
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT products_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE tombstones
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT tombstones_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE vendors
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT vendors_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

COMMIT;
