BEGIN;

ALTER TABLE cert_subject_mapping
    ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMP;

COMMIT;
