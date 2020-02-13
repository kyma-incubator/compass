ALTER TABLE api_definitions ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE api_definitions
    DROP CONSTRAINT api_definitions_package_id_fk;
ALTER TABLE api_definitions DROP COLUMN package_id;

ALTER TABLE event_api_definitions ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE event_api_definitions
    DROP CONSTRAINT event_api_definitions_package_id_fk;
ALTER TABLE event_api_definitions DROP COLUMN package_id;

ALTER TABLE documents ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE documents
    DROP CONSTRAINT documents_package_id_fk;
ALTER TABLE documents DROP COLUMN package_id;

DROP TABLE package_instance_auths;
DROP TABLE packages;

DROP TYPE package_instance_auth_status_condition;
