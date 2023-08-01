BEGIN;

CREATE OR REPLACE FUNCTION delete_cert_subject_mapping_on_app_template_delete()
RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM cert_subject_mapping
    WHERE internal_consumer_id = OLD.id::varchar(256);
    RETURN OLD;
END;
$$
LANGUAGE plpgsql;


CREATE TRIGGER trigger_delete_cert_subject_mapping_on_app_template_delete
AFTER DELETE ON app_templates
FOR EACH ROW
EXECUTE FUNCTION delete_cert_subject_mapping_on_app_template_delete();
COMMIT;
