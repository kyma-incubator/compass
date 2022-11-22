BEGIN;

DROP INDEX fetch_request_specification_id;
DROP INDEX api_specifications_tenants_app_id;
DROP INDEX event_specifications_tenants_app_id;

CREATE INDEX event_specifications_tenants_app_id ON specifications(event_def_id);

CREATE INDEX api_specifications_tenants_app_id ON specifications(api_def_id);

CREATE INDEX fetch_request_specification_id ON fetch_requests(spec_id);

CREATE INDEX bundle_references_api_definition_id ON bundle_references(api_def_id);
CREATE INDEX bundle_references_event_definition_id ON bundle_references(event_def_id);
CREATE INDEX bundle_references_bundle_id ON bundle_references(bundle_id);

CREATE INDEX labels_application_template_id on labels(app_template_id) WHERE labels.app_template_id IS NOT NULL;

COMMIT;
