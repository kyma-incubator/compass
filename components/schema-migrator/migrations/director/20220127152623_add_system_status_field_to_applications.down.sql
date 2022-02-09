BEGIN;

ALTER TABLE applications
DROP CONSTRAINT system_status_len;

ALTER TABLE applications
DROP COLUMN system_status;

COMMIT;
