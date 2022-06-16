BEGIN;

ALTER TABLE app_templates DROP COLUMN application_namespace;

DELETE FROM labels
WHERE app_id IS NOT NULL AND app_id IN (SELECT app_id FROM labels WHERE key = 'applicationType' AND value = '"SAP S/4HANA Cloud"') AND key = 'region' AND value = '"default"';

COMMIT;
