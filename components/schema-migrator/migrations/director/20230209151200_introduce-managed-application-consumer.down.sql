BEGIN;

ALTER TABLE cert_subject_mapping
    ADD CONSTRAINT consumer_type CHECK ( consumer_type IN ('Runtime', 'Integration System', 'Application', 'Super Admin"', 'Business Integration', 'Technical Client', 'global'));

COMMIT;
