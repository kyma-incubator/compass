BEGIN;

CREATE INDEX IF NOT EXISTS labels_app_region_idx ON labels (app_id, value) WHERE key = 'region';
CREATE INDEX IF NOT EXISTS labels_app_tmpl_region_idx ON labels (app_template_id, value) WHERE key = 'region';

DROP VIEW IF EXISTS tenants_apps;

CREATE OR REPLACE VIEW tenants_apps
            (tenant_id, formation_id, id, name, description, status_condition, status_timestamp, healthcheck_url,
             integration_system_id, provider_name, base_url, labels, tags, ready, created_at, updated_at, deleted_at,
             error, app_template_id, correlation_ids, system_number, application_namespace, region, local_tenant_id,
             product_type, fa_formation_id, assignment_id, formation_type_id, target_id)
AS
SELECT DISTINCT t_apps.tenant_id,
                t_apps.formation_id,
                apps.id,
                apps.name,
                apps.description,
                apps.status_condition,
                apps.status_timestamp,
                apps.healthcheck_url,
                apps.integration_system_id,
                apps.provider_name,
                apps.base_url,
                apps.labels,
                apps.tags,
                apps.ready,
                apps.created_at,
                apps.updated_at,
                apps.deleted_at,
                apps.error,
                apps.app_template_id,
                apps.correlation_ids,
                apps.system_number,
                COALESCE(apps.application_namespace, tmpl.application_namespace) AS application_namespace,
                COALESCE(labels_app.value, labels_tmpl.value) AS region,
                apps.local_tenant_id,
                tmpl.name                                                        AS product_type,
                formation_details.formation_id,
                formation_details.assignment_id,
                formation_details.formation_type_id,
                COALESCE(formation_details.target_id, 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'::uuid)
FROM applications apps
         LEFT JOIN app_templates tmpl ON apps.app_template_id = tmpl.id
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
               FROM apps_subaccounts) t_apps ON apps.id = t_apps.id
         LEFT JOIN (SELECT DISTINCT fa.id AS assignment_id,
                                    fa.formation_id,
                                    f.formation_template_id AS formation_type_id,
                                    fa.source,
                                    fa.target AS target_id
                    FROM formation_assignments fa JOIN formations f ON fa.formation_id = f.id) formation_details ON formation_details.source = t_apps.id AND formation_details.formation_id = t_apps.formation_id -- the second join by formation_id is required so that we have records (tenant + default formation_id + null formation_details) that are needed for backwards compatibility when we call the ORD Service only with tenant filtering; otherwise, since the join will be only on source, there are records (tenant + default formation_id + valid formation_details) which are invalid because we want null ones as there is no formation filtering
         LEFT JOIN labels AS labels_app ON labels_app.app_id = apps.id AND labels_app.key = 'region'
         LEFT JOIN labels AS labels_tmpl ON labels_tmpl.app_template_id = tmpl.id AND labels_tmpl.key = 'region';
COMMIT;