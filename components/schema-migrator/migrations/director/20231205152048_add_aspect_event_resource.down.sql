BEGIN;

-- Drop views --
DROP VIEW IF EXISTS aspect_event_resources_tenants;
DROP VIEW IF EXISTS tenants_aspect_event_resources;

-- Drop index for aspect_event_resources table
DROP INDEX IF EXISTS aspect_event_resources_app_id;

-- Drop table aspect_event_resources
    DROP TABLE IF EXISTS aspect_event_resources;
COMMIT;
