BEGIN;

DROP VIEW api_definition_successors;
DROP VIEW event_api_definition_successors;

ALTER TABLE api_definitions
    ADD COLUMN successor VARCHAR(256),
    DROP COLUMN successors;

ALTER TABLE event_api_definitions
    ADD COLUMN successor VARCHAR(256),
    DROP COLUMN successors;

ALTER TYPE policy_level RENAME TO policy_level_old;

CREATE TYPE policy_level AS ENUM ('sap', 'sap-partner','custom');

ALTER TABLE packages ALTER COLUMN policy_level TYPE TEXT;

UPDATE packages SET policy_level = 'sap' WHERE policy_level = 'sap:core:v1';
UPDATE packages SET policy_level = 'sap-partner' WHERE policy_level = 'sap:partner:v1';

ALTER TABLE packages
    ALTER COLUMN policy_level TYPE policy_level
    USING policy_level::text::policy_level;

DROP TYPE policy_level_old;

COMMIT;
