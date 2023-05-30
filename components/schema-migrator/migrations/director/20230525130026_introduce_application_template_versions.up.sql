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
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
    ADD CONSTRAINT api_definitions_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id);

ALTER TABLE bundles
    ADD COLUMN app_template_version_id UUID,
    ALTER COLUMN app_id DROP NOT NULL,
    ADD CONSTRAINT bundles_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
    ADD CONSTRAINT bundles_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id);

ALTER TABLE documents
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT documents_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE event_api_definitions
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT event_api_definitions_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
    ADD CONSTRAINT event_api_definitions_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id);

ALTER TABLE packages
    ADD COLUMN app_template_version_id UUID,
    ALTER COLUMN app_id DROP NOT NULL,
    ADD CONSTRAINT packages_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
    ADD CONSTRAINT packages_app_template_version_id_ord_id_unique UNIQUE (app_template_version_id, ord_id);

ALTER TABLE products
    ADD COLUMN app_template_version_id UUID,
    ALTER COLUMN app_id DROP NOT NULL,
    ADD CONSTRAINT products_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE tombstones
    ADD COLUMN app_template_version_id UUID,
    ALTER COLUMN app_id DROP NOT NULL,
    ADD CONSTRAINT tombstones_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;

ALTER TABLE vendors
    ADD COLUMN app_template_version_id UUID,
    ADD CONSTRAINT vendors_app_template_version_id_fk
        FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE ;


CREATE INDEX api_definitions_app_template_version_id ON api_definitions (app_template_version_id);

CREATE INDEX bundles_app_template_version_id ON bundles (app_template_version_id);

CREATE INDEX documents_app_template_version_id ON documents (app_template_version_id);

CREATE INDEX event_api_definitions_app_template_version_id ON event_api_definitions (app_template_version_id);

CREATE INDEX packages_app_template_version_id ON packages (app_template_version_id);

CREATE INDEX products_app_template_version_id ON products (app_template_version_id);

CREATE INDEX tombstones_app_template_version_id ON tombstones (app_template_version_id);

CREATE INDEX vendors_app_template_version_id ON vendors (app_template_version_id);

COMMIT;
