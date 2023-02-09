BEGIN;

ALTER TABLE webhooks
    ADD COLUMN parameters JSONB;

COMMIT;
