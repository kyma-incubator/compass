BEGIN;

ALTER TABLE cert_subject_mapping
    DROP CONSTRAINT cert_subject_mapping_consumer_type_check,
    DROP CONSTRAINT cert_subject_mapping_internal_consumer_id_check;

ALTER TABLE cert_subject_mapping
    ALTER COLUMN internal_consumer_id TYPE VARCHAR(256);

COMMIT;
