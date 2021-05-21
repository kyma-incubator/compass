BEGIN;

ALTER TABLE bundles DROP CONSTRAINT bundles_applications_ready_fk;
ALTER TABLE bundles
    ADD CONSTRAINT bundles_applications_ready_fk
        FOREIGN KEY (app_id, ready) REFERENCES applications (id, ready) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE api_definitions DROP CONSTRAINT api_definitions_bundles_ready_fk;
ALTER TABLE api_definitions
    ADD CONSTRAINT api_definitions_bundles_ready_fk
        FOREIGN KEY (bundle_id, ready) REFERENCES bundles (id, ready) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE event_api_definitions DROP CONSTRAINT event_api_definitions_bundles_ready_fk;
ALTER TABLE event_api_definitions
    ADD CONSTRAINT event_api_definitions_bundles_ready_fk
        FOREIGN KEY (bundle_id, ready) REFERENCES bundles (id, ready) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE documents DROP CONSTRAINT documents_bundles_ready_fk;
ALTER TABLE documents
    ADD CONSTRAINT documents_bundles_ready_fk
        FOREIGN KEY (bundle_id, ready) REFERENCES bundles (id, ready) ON DELETE CASCADE ON UPDATE CASCADE;

COMMIT;
