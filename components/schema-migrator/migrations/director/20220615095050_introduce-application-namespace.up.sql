BEGIN;

ALTER TABLE app_templates ADD COLUMN application_namespace VARCHAR(256);

INSERT INTO labels(id, tenant_id, app_id, runtime_id, key, value, runtime_context_id, version, app_template_id)
SELECT uuid_generate_v4(), NULL::UUID, app_id, NULL::UUID, 'region', '"default"', NULL::UUID, 0, NULL::UUID
FROM labels
WHERE key = 'applicationType' AND value = '"SAP S/4HANA Cloud"';

COMMIT;
