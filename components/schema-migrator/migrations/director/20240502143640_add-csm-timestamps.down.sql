BEGIN;

ALTER table cert_subject_mapping
    DROP created_at,
    DROP updated_at;

COMMIT;
