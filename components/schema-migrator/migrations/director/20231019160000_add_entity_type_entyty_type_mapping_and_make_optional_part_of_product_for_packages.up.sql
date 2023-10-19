BEGIN;

CREATE TABLE entity_types
(
    id                          UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    ready                       BOOLEAN DEFAULT TRUE,
    created_at                  TIMESTAMP,
    updated_at                  TIMESTAMP,
    deleted_at                  TIMESTAMP, 
    error                       JSONB, 
    app_id                      UUID,
    app_template_version_id     UUID,
    ord_id                      VARCHAR(256) NOT NULL,
    local_id                    VARCHAR(256) NOT NULL,
    correlation_ids             JSONB,
    level                       VARCHAR(256) NOT NULL,
    title                       VARCHAR(256) NOT NULL,
    short_description           VARCHAR(256),
    description                 TEXT,
    system_instance_aware       BOOLEAN,
    changelog_entries           JSONB,
    package_id                  VARCHAR(256) NOT NULL,
    visibility                  VARCHAR(256) NOT NULL,
    links                       JSONB,
    part_of_products            JSONB,
    policy_level                policy_level,
    custom_policy_level         VARCHAR(256),
    release_status              release_status,
    sunset_date                 VARCHAR(256),
    successors                  JSONB,
    extensible                  JSONB,
    tags                        JSONB,
    labels                      JSONB,
    documentation_labels        JSONB,
    resource_hash               VARCHAR(255),
    version_value               VARCHAR(256),
    version_deprecated          BOOLEAN,
    version_deprecated_since    VARCHAR(256),
    version_for_removal         BOOLEAN
);


DROP VIEW IF EXISTS correlation_ids_entity_types;

CREATE VIEW correlation_ids_entity_types AS
SELECT id                  AS entity_type_id,
       elements.value      AS value
FROM entity_types,
     jsonb_array_elements_text(entity_types.correlation_ids) AS elements;


DROP VIEW IF EXISTS changelog_entries_entity_types;

CREATE VIEW changelog_entries_entity_types AS
SELECT id                      AS entity_type_id,
       entries.version         AS version,
       entries."releaseStatus" AS release_status,
       entries.date            AS date,
       entries.description     AS description,
       entries.url             AS url
FROM entity_types,
     jsonb_to_recordset(entity_types.changelog_entries) AS entries(version TEXT, "releaseStatus" TEXT, date TEXT, description TEXT, url TEXT);


DROP VIEW IF EXISTS links_entity_types;

CREATE VIEW links_entity_types AS
SELECT id                AS entity_type_id,
       links.title       AS title,
       links.url         AS url,
       links.description AS description
FROM entity_types,
     jsonb_to_recordset(entity_types.links) AS links(title TEXT, description TEXT, url TEXT);


DROP VIEW IF EXISTS entity_type_product;

CREATE VIEW entity_type_product AS
SELECT entity_types.id  AS entity_type_id,
       p.id             AS product_id
FROM entity_types,
     jsonb_array_elements_text(entity_types.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = entity_types.app_id OR p.app_id IS NULL;

DROP VIEW IF EXISTS entity_type_successors;

CREATE VIEW entity_type_successors AS
SELECT id             AS entity_type_id,
       elements.value AS value
FROM entity_types,
     jsonb_array_elements_text(entity_types.successors) AS elements;


DROP VIEW IF EXISTS entity_type_extensible;

CREATE VIEW entity_type_extensible AS
SELECT id                  AS entity_type_id,
       actions.supported   AS supported,
       actions.description AS description
FROM entity_types,
     jsonb_to_record(entity_types.extensible) AS actions(supported TEXT, description TEXT);


DROP VIEW IF EXISTS ord_tags_entity_types;

CREATE VIEW ord_tags_entity_types AS
SELECT id                  AS entity_type_id,
       elements.value      AS value
FROM entity_types,
     jsonb_array_elements_text(entity_types.tags) AS elements;


DROP VIEW IF EXISTS ord_labels_entity_types;

CREATE VIEW ord_labels_entity_types AS
SELECT id                  AS entity_type_id,
       expand.key          AS key,
       elements.value      AS value
FROM entity_types,
     jsonb_each(entity_types.labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;


DROP VIEW IF EXISTS ord_documentation_labels_entity_types;

CREATE VIEW ord_documentation_labels_entity_types AS
SELECT id                  AS entity_type_id,
       expand.key          AS key,
       elements.value      AS value
FROM entity_types,
     jsonb_each(entity_types.documentation_labels) AS expand,
     jsonb_array_elements_text(expand.value) AS elements;


DROP VIEW IF EXISTS entity_types_tenants;

CREATE VIEW entity_types_tenants AS
SELECT e.*, ta.tenant_id, ta.owner FROM entity_types AS e
                                            INNER JOIN tenant_applications ta ON ta.id = e.app_id;

DROP VIEW IF EXISTS tenants_entity_types;

CREATE OR REPLACE VIEW tenants_entity_types
            (tenant_id, formation_id, id, ord_id, app_id, local_id, level, title, short_description, description, system_instance_aware, 
            changelog_entries, package_id, visibility, links, part_of_products, policy_level,
            custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, 
            documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                et.id,
                et.ord_id,
                et.app_id,
                et.local_id,
                et.level,
                et.title,
                et.short_description,
                et.description,
                et.system_instance_aware,
                et.changelog_entries,
                et.package_id,
                et.visibility,
                et.links, 
                et.part_of_products,
                et.policy_level,
                et.custom_policy_level,
                et.release_status,
                et.sunset_date,
                et.successors,
                et.extensible,
                et.tags,
                et.labels,
                et.documentation_labels,
                et.resource_hash,
                et.version_value,
                et.version_deprecated,
                et.version_deprecated_since,
                et.version_for_removal
FROM entity_types et
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
              ON et.app_id = t_apps.id,
              jsonb_to_record(et.extensible) actions(supported text, description text);

CREATE INDEX entity_types_app_template_version_id ON entity_types (app_template_version_id);


ALTER TABLE packages
    ALTER COLUMN part_of_products DROP NOT NULL;

COMMIT;
