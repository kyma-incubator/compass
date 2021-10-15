BEGIN;

ALTER TABLE webhooks DROP CONSTRAINT webhook_access_strategy_ord_only;
ALTER TABLE webhooks DROP COLUMN access_strategy;

COMMIT;
