BEGIN;

-- Drop views --
DROP VIEW IF EXISTS aspect_event_resources_subset;
DROP VIEW IF EXISTS aspect_event_resources;
DROP VIEW IF EXISTS aspects_tenants; -- this view will not be changed but it depends on the column aspects.event_resources; that's why the view has to be dropped and re-created in order to drop the column
DROP VIEW IF EXISTS tenants_aspects;

-- Create aspect_event_resources table
CREATE TABLE aspect_event_resources (
                         id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
                         aspect_id UUID NOT NULL,
                         CONSTRAINT aspect_event_resources_aspect_id_fkey FOREIGN KEY (aspect_id) REFERENCES aspects (id) ON DELETE CASCADE,
                         app_id UUID,
                         CONSTRAINT aspect_event_resources_app_id_fkey FOREIGN KEY (app_id) REFERENCES applications (id) ON DELETE CASCADE,
                         app_template_version_id UUID,
                         CONSTRAINT aspect_event_resources_app_template_version_id_fk FOREIGN KEY (app_template_version_id) REFERENCES app_template_versions (id) ON DELETE CASCADE,
                         ord_id VARCHAR(256) NOT NULL,
                         CONSTRAINT aspect_event_resources_ord_id_unique UNIQUE (app_id, aspect_id, ord_id),
                         min_version VARCHAR(256),
                         subset JSONB,
                         ready BOOLEAN DEFAULT TRUE,
                         CONSTRAINT aspect_event_resource_id_ready_unique UNIQUE (id, ready),
                         created_at TIMESTAMP,
                         updated_at TIMESTAMP,
                         deleted_at TIMESTAMP,
                         error JSONB
);

-- a manual migration of this column will be performed that's why we drop it directly
ALTER TABLE aspects
    DROP COLUMN event_resources;

-- Create index for aspect_event_resources table
CREATE INDEX IF NOT EXISTS aspect_event_resources_app_id ON aspect_event_resources (app_id);

-- Create view tenants_aspect_event_resources
CREATE VIEW tenants_aspect_event_resources
            (tenant_id, formation_id, id, aspect_id, app_id, ord_id, min_version, subset, ready, created_at, updated_at, deleted_at, error)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                a.id,
                a.aspect_id,
                a.app_id,
                a.ord_id,
                a.min_version,
                a.subset,
                a.ready,
                a.created_at,
                a.updated_at,
                a.deleted_at,
                a.error
FROM aspect_event_resources a
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON a.app_id = t_apps.id;


CREATE VIEW aspect_event_resources_subset AS
SELECT
    id                                  AS event_resource_id,
    entries."eventType"                 AS event_type
FROM aspect_event_resources a,
     jsonb_to_recordset(a.subset) AS entries("eventType" TEXT); -- if the JSON key has a capital letter then it should be defined with "" in the AS clause

-- Create aspect_event_resources_tenants view
CREATE VIEW aspect_event_resources_tenants AS
SELECT a.*, ta.tenant_id, ta.owner FROM aspect_event_resources AS a
                                            INNER JOIN tenant_applications ta ON ta.id = a.app_id;

-- Recreate views
CREATE VIEW aspects_tenants AS
SELECT a.*, ta.tenant_id, ta.owner FROM aspects AS a
                                            INNER JOIN tenant_applications ta ON ta.id = a.app_id;

-- event_resources column is removed
CREATE OR REPLACE VIEW tenants_aspects
            (tenant_id, formation_id, id, integration_dependency_id, app_id, title, description, mandatory,
             support_multiple_providers, api_resources, ready, created_at, updated_at, deleted_at,
             error)
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
                a.ready,
                a.created_at,
                a.updated_at,
                a.deleted_at,
                a.error
FROM aspects a
         JOIN (SELECT a1.id,
                      a1.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM tenant_applications a1
               UNION ALL
               SELECT af.app_id,
                      'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid AS tenant_id,
                      af.formation_id
               FROM apps_formations_id af
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id,
                      'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
               FROM apps_subaccounts) t_apps ON a.app_id = t_apps.id;

COMMIT;
