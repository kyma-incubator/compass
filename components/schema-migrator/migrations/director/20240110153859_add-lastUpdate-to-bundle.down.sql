BEGIN;
DROP VIEW IF EXISTS tenants_bundles;

-- Alter table - remove `last_update` from Consumption Bundle
ALTER TABLE bundles
    DROP COLUMN last_update;

CREATE OR REPLACE VIEW tenants_bundles
            (tenant_id, formation_id, id, app_id, name, description, version, instance_auth_request_json_schema,
             default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, tags,
             credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids,
             resource_hash)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                b.id,
                b.app_id,
                b.name,
                b.description,
                b.version,
                b.instance_auth_request_json_schema,
                b.default_instance_auth,
                b.ord_id,
                b.local_tenant_id,
                b.short_description,
                b.links,
                b.labels,
                b.tags,
                b.credential_exchange_strategies,
                b.ready,
                b.created_at,
                b.updated_at,
                b.deleted_at,
                b.error,
                b.correlation_ids,
                b.resource_hash
FROM bundles b
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
               FROM apps_subaccounts) t_apps ON b.app_id = t_apps.id;
COMMIT;
