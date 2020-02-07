ALTER TABLE api_definitions DROP COLUMN package_definition_id;
ALTER TABLE api_definitions
    DROP CONSTRAINT api_definitions_package_definition_id_fk;

ALTER TABLE event_api_definitions DROP COLUMN package_definition_id;
ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_api_definitions_package_definition_id_fk;

ALTER TABLE documents DROP COLUMN package_definition_id;
ALTER TABLE documents
    DROP CONSTRAINT documents_package_definition_id_fk;

DROP TABLE api_instance_auths;
DROP TABLE package_definitions;
