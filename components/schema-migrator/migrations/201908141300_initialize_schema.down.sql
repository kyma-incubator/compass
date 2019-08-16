-- Label

DROP TABLE labels;

-- Label Definition

DROP TABLE label_definitions;

-- Runtime Auths

DROP TABLE runtime_auths;

-- Runtime

DROP TABLE runtimes;
DROP TYPE runtime_status_condition;

-- Fetch Request

ALTER TABLE api_definitions
    DROP COLUMN fetch_request_id;
ALTER TABLE documents
    DROP COLUMN fetch_request_id;
ALTER TABLE event_api_definitions
    DROP COLUMN fetch_request_id;

DROP TABLE fetch_requests;
DROP TYPE fetch_request_mode;
DROP TYPE fetch_request_status_condition;

-- Webhooks

DROP TABLE webhooks;
DROP TYPE webhook_type;

-- API Definitions

DROP TABLE api_definitions;
DROP TYPE api_spec_format;
DROP TYPE api_spec_type;

-- Event API Definitions

DROP TABLE event_api_definitions;
DROP TYPE event_api_spec_format;
DROP TYPE event_api_spec_type;

-- Documents

DROP TABLE documents;
DROP TYPE document_format;

-- Application

DROP TABLE applications;
DROP TYPE application_status_condition;
