BEGIN;

ALTER TABLE fetch_requests
    ADD COLUMN status_message varchar(256);

COMMIT;
