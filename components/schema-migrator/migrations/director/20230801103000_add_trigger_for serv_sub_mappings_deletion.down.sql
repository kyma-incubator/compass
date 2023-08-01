BEGIN;

DROP TRIGGER IF EXISTS trigger_delete_cert_subject_mapping_on_app_template_delete ON app_templates;
DROP FUNCTION IF EXISTS delete_cert_subject_mapping_on_app_template_delete();

COMMIT;