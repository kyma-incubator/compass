BEGIN;

UPDATE api_definitions
SET target_urls='[]'
WHERE api_definitions.target_urls IS NULL;

ALTER TABLE api_definitions
    ALTER COLUMN target_urls SET NOT NULL;

COMMIT;
