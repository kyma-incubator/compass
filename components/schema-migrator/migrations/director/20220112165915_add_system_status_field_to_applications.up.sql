BEGIN;

ALTER TABLE applications
    ADD COLUMN system_status TEXT;

COMMIT;
