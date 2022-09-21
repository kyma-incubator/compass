BEGIN;
DROP VIEW tenants_bundles;
DROP VIEW bundles_tenants;

ALTER TABLE bundles
    ALTER COLUMN short_description TYPE varchar(2048);

CREATE OR REPLACE VIEW bundles_tenants AS
SELECT b.*, ta.tenant_id, ta.owner FROM bundles AS b
                                            INNER JOIN tenant_applications ta ON ta.id = b.app_id;

CREATE OR REPLACE VIEW tenants_bundles
            (tenant_id, id, app_id, name, description, instance_auth_request_json_schema,
             default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready,
             created_at, updated_at, deleted_at, error, correlation_ids)
AS
SELECT DISTINCT t_apps.tenant_id,
                b.id,
                b.app_id,
                b.name,
                b.description,
                b.instance_auth_request_json_schema,
                b.default_instance_auth,
                b.ord_id,
                b.short_description,
                b.links,
                b.labels,
                b.credential_exchange_strategies,
                b.ready,
                b.created_at,
                b.updated_at,
                b.deleted_at,
                b.error,
                b.correlation_ids
FROM bundles b
         JOIN (SELECT a1.id,
                      a1.tenant_id AS tenant_id
               FROM tenant_applications a1
               UNION ALL
               SELECT apps_subaccounts.id,
                      apps_subaccounts.tenant_id
               FROM apps_subaccounts) t_apps
              ON b.app_id = t_apps.id;


COMMIT;
