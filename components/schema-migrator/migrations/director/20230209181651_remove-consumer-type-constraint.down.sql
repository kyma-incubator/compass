BEGIN;

ALTER TABLE cert_subject_mapping
    ADD CONSTRAINT cert_subject_mapping_consumer_type_check CHECK ( consumer_type IN ('Runtime', 'Integration System', 'Application', 'Super Admin', 'Business Integration', 'Technical Client', 'global'));

COMMIT;
