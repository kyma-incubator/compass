BEGIN;

ALTER TABLE cert_subject_mapping
    DROP CONSTRAINT consumer_type;

COMMIT;
