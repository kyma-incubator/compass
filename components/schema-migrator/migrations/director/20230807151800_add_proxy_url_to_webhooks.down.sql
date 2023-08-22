BEGIN;

ALTER TABLE webhooks
    DROP COLUMN proxy_url;

COMMIT;