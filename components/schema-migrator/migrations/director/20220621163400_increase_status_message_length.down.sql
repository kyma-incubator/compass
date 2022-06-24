BEGIN;

DROP VIEW IF EXISTS event_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS api_specifications_fetch_requests_tenants;
DROP VIEW IF EXISTS document_fetch_requests_tenants;

ALTER TABLE fetch_requests
    ALTER COLUMN "status_message" TYPE varchar(256);

CREATE OR REPLACE VIEW document_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                             INNER JOIN documents d ON fr.document_id = d.id
                                             INNER JOIN tenant_applications ta ON ta.id = d.app_id;

CREATE OR REPLACE VIEW api_specifications_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                             INNER JOIN specifications s ON fr.spec_id = s.id
                                             INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ad.app_id;

CREATE OR REPLACE VIEW event_specifications_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                             INNER JOIN specifications s ON fr.spec_id = s.id
                                             INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ead.app_id;

COMMIT;
