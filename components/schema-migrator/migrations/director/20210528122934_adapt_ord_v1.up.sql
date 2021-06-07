BEGIN;

ALTER TABLE api_definitions
    ALTER COLUMN target_urls DROP NOT NULL;

COMMIT;
