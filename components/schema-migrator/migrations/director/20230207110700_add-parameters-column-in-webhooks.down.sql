BEGIN;

ALTER TABLE webhooks
    DROP COLUMN parameters;

COMMIT;