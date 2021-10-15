BEGIN;

ALTER TABLE webhooks ADD COLUMN access_strategy VARCHAR(256);
ALTER TABLE webhooks ADD CONSTRAINT webhook_access_strategy_ord_only
    CHECK ((webhooks.access_strategy IS NOT NULL AND webhooks.type = 'OPEN_RESOURCE_DISCOVERY') OR webhooks.access_strategy IS NULL);

COMMIT;
