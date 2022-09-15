BEGIN;

ALTER TABLE webhooks
    DROP COLUMN created_at;

COMMIT;
