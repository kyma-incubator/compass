BEGIN;

-- Drop new views
DROP VIEW IF EXISTS tenants_aspect_event_resources;
DROP VIEW IF EXISTS aspect_event_resources_subset;
DROP VIEW IF EXISTS aspect_event_resources_tenants;
DROP VIEW IF EXISTS aspects_tenants;
DROP VIEW IF EXISTS tenants_aspects;

-- Drop index for aspect_event_resources table
DROP INDEX IF EXISTS aspect_event_resources_app_id;

-- Drop table aspect_event_resources
DROP TABLE IF EXISTS aspect_event_resources;

-- Add column back
ALTER TABLE aspects
    ADD COLUMN event_resources JSONB;

-- Recreate old views
CREATE VIEW aspect_event_resources AS
SELECT asp.id                               AS aspect_id,
       events.id                            AS event_resource_id,
       entries."ordId"                      AS ord_id,
       entries."minVersion"                 AS min_version,
       entries.subset                       AS subset
FROM aspects asp JOIN event_api_definitions events ON asp.app_id = events.app_id, jsonb_to_recordset(asp.event_resources) AS entries("ordId" TEXT, "minVersion" TEXT, subset JSONB)
WHERE events.ord_id = entries."ordId";

CREATE VIEW aspect_event_resources_subset AS
SELECT
    a.event_resource_id                 AS event_resource_id,
    entries."eventType"                 AS event_type
FROM aspect_event_resources a,
     jsonb_to_recordset(a.subset) AS entries("eventType" TEXT); -- if the JSON key has a capital letter then it should be defined with "" in the AS clause

-- Recreate views
CREATE VIEW aspects_tenants AS
SELECT a.*, ta.tenant_id, ta.owner FROM aspects AS a
                                            INNER JOIN tenant_applications ta ON ta.id = a.app_id;

CREATE VIEW tenants_aspects
            (tenant_id, formation_id, id, integration_dependency_id, app_id, title, description, mandatory, support_multiple_providers, api_resources, event_resources, ready, created_at,
             updated_at, deleted_at, error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                a.id,
                a.integration_dependency_id,
                a.app_id,
                a.title,
                a.description,
                a.mandatory,
                a.support_multiple_providers,
                a.api_resources,
                a.event_resources,
                a.ready,
                a.created_at,
                a.updated_at,
                a.deleted_at,
                a.error
FROM aspects a
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      apps_subaccounts.formation_id
               FROM apps_subaccounts
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS formation_id
               FROM apps_subaccounts) t_apps ON a.app_id = t_apps.id;


COMMIT;
