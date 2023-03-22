BEGIN;

CREATE TABLE data_products (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id uuid NOT NULL,
    CONSTRAINT data_products_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    ord_id VARCHAR(256) NOT NULL,
    local_id VARCHAR(256),
    title VARCHAR(256),
    short_description VARCHAR(256),
    description TEXT,
    version VARCHAR(256),
    release_status release_status,
    visibility TEXT,
    CONSTRAINT data_product_visibility_check CHECK (visibility in ('public', 'internal', 'private')),
    package_id uuid NOT NULL,
    CONSTRAINT data_products_package_id_fk FOREIGN KEY (package_id) REFERENCES packages(id),
    tags JSONB,
    industry JSONB,
    line_of_business JSONB,
    product_type VARCHAR(256),
    data_product_owner VARCHAR(256)
);

ALTER TABLE bundle_references
    ADD COLUMN data_product_id uuid;

ALTER TABLE bundle_references
    ADD CONSTRAINT bundle_references_data_product_id_fkey FOREIGN KEY (data_product_id) REFERENCES data_products(id);

ALTER TABLE bundle_references
    DROP CONSTRAINT valid_refs;

CREATE TABLE ports (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    data_product_id uuid NOT NULL,
    CONSTRAINT ports_data_product_id_fk FOREIGN KEY (data_product_id) REFERENCES data_products (id) ON DELETE CASCADE,
    app_id uuid NOT NULL,
    CONSTRAINT ports_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    name VARCHAR(256),
    port_type VARCHAR(256),
    description TEXT,
    producer_cardinality VARCHAR(256),
    disabled bool DEFAULT FALSE
);

CREATE TABLE port_api_reference (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id uuid NOT NULL,
    CONSTRAINT port_api_ref_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    port_id uuid NOT NULL,
    CONSTRAINT port_api_reference_port_id_fk FOREIGN KEY (port_id) REFERENCES ports (id) ON DELETE CASCADE,
    api_id uuid NOT NULL,
    CONSTRAINT port_api_reference_api_id_fk FOREIGN KEY (api_id) REFERENCES api_definitions (id)
);

CREATE TABLE port_event_reference (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    app_id uuid NOT NULL,
    CONSTRAINT port_event_ref_application_id_fk FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    port_id uuid NOT NULL,
    CONSTRAINT port_event_reference_port_id_fk FOREIGN KEY (port_id) REFERENCES ports (id) ON DELETE CASCADE,
    event_id uuid NOT NULL,
    CONSTRAINT port_event_reference_event_id_fk FOREIGN KEY (event_id) REFERENCES event_api_definitions (id),
    min_version VARCHAR(256)
);

CREATE OR REPLACE VIEW data_product_tenants AS
SELECT dp.*, ta.tenant_id, ta.owner FROM data_products AS dp
                                             INNER JOIN tenant_applications ta ON ta.id = dp.app_id;

CREATE OR REPLACE VIEW tenants_data_products
            (tenant_id, formation_id, id, app_id, ord_id, local_id, title, short_description, description,
             version, release_status, visibility, package_id, tags, industry, line_of_business, product_type, data_product_owner)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                dp.id,
                dp.app_id,
                dp.ord_id,
                dp.local_id,
                dp.title,
                dp.short_description,
                dp.description,
                dp.version,
                dp.release_status,
                dp.visibility,
                dp.package_id,
                dp.tags,
                dp.industry,
                dp.line_of_business,
                dp.product_type,
                dp.data_product_owner
FROM data_products dp
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      apps_subaccounts.formation_id
               FROM apps_subaccounts
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps
              ON dp.app_id = t_apps.id;


CREATE VIEW ord_data_products_tags AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.tags) AS elements;

CREATE VIEW ord_data_products_industry AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.industry) AS elements;

CREATE VIEW ord_data_products_line_of_business AS
SELECT id                  AS data_product_id,
       elements.value      AS value
FROM data_products,
     jsonb_array_elements_text(data_products.line_of_business) AS elements;


CREATE OR REPLACE VIEW tenants_ports
            (tenant_id, formation_id, id, data_product_id, app_id, name, port_type, description,
             producer_cardinality, disabled)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                p.id,
                p.data_product_id,
                p.app_id,
                p.name,
                p.port_type,
                p.description,
                p.producer_cardinality,
                p.disabled
FROM ports p
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      apps_subaccounts.formation_id
               FROM apps_subaccounts
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps
              ON p.app_id = t_apps.id;


COMMIT;
