BEGIN;

ALTER TABLE packages
    ALTER COLUMN part_of_products SET NOT NULL;

DROP INDEX IF EXISTS entity_types_app_template_version_id;

DROP VIEW IF EXISTS tenants_entity_types;
DROP VIEW IF EXISTS correlation_ids_entity_types;
DROP VIEW IF EXISTS changelog_entries_entity_types;
DROP VIEW IF EXISTS links_entity_types;
DROP VIEW IF EXISTS entity_type_product;
DROP VIEW IF EXISTS entity_type_successors;
DROP VIEW IF EXISTS entity_type_extensible;
DROP VIEW IF EXISTS ord_tags_entity_types;
DROP VIEW IF EXISTS ord_labels_entity_types;
DROP VIEW IF EXISTS ord_documentation_labels_entity_types;

DROP TABLE entity_types;

COMMIT;
