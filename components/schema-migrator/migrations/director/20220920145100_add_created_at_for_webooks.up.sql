BEGIN;

ALTER TABLE webhooks
    ADD COLUMN created_at timestamp;

COMMIT;
