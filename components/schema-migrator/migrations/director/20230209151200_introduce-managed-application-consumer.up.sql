BEGIN;

ALTER TABLE cert_subject_mapping
    DROP CONSTRAINT cert_subject_mapping_consumer_type_check;

COMMIT;
