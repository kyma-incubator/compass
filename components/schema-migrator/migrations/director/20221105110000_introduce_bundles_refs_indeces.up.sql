BEGIN;

DROP INDEX fetch_request_specification_id;

CREATE INDEX fetch_request_specification_id ON fetch_requests(spec_id) WHERE fetch_requests.spec_id IS NOT NULL;

CREATE INDEX bundle_references_api_definition_id ON bundle_references(api_def_id) WHERE bundle_references.api_def_id IS NOT NULL;
CREATE INDEX bundle_references_event_definition_id ON bundle_references(event_def_id) WHERE bundle_references.event_def_id IS NOT NULL;
CREATE INDEX bundle_references_bundle_id ON bundle_references(bundle_id);

COMMIT;
