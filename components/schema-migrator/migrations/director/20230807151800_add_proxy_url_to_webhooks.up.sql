BEGIN;

ALTER TABLE webhooks
    ADD COLUMN proxy_url varchar(256);

COMMIT;