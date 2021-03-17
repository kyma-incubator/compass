BEGIN;

ALTER TABLE api_definitions
    DROP COLUMN implementation_standard,
    DROP COLUMN custom_implementation_standard,
    DROP COLUMN custom_implementation_standard_description;

COMMIT;
