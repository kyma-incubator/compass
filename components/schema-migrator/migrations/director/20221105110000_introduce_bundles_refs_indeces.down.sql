BEGIN;

DROP INDEX bundle_references_api_definition_id;
DROP INDEX bundle_references_event_definition_id;
DROP INDEX bundle_references_bundle_id;
DROP INDEX labels_application_template_id;

COMMIT;
