BEGIN;

ALTER TABLE applications
    DROP COLUMN system_status;

COMMIT;
