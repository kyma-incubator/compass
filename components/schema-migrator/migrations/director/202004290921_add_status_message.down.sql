BEGIN;

ALTER TABLE fetch_requests
    DROP COLUMN status_message;

COMMIT;
