BEGIN;

-- create view for apis based on bundle_reference table including the formation as well
CREATE OR REPLACE VIEW tenants_api_bundle_reference AS
SELECT  api_def_id as api_definition_id,
        bundle_id,
        formation_apps.formation_id
FROM bundle_references br JOIN bundles b ON br.bundle_id = b.id JOIN (SELECT a1.id,
                                                                             'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                        FROM applications a1
                                                                        UNION ALL
                                                                        SELECT af.app_id,
                                                                               af.formation_id
                                                                        FROM apps_formations_id af) formation_apps ON b.app_id = formation_apps.id
WHERE api_def_id IS NOT NULL;

-- create view for events based on bundle_reference table including the formation as well
CREATE OR REPLACE VIEW tenants_event_bundle_reference AS
SELECT  event_def_id,
        bundle_id,
        formation_apps.formation_id
FROM bundle_references br JOIN bundles b ON br.bundle_id = b.id JOIN (SELECT apps.id,
                                                                             'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                        FROM applications apps
                                                                        UNION ALL
                                                                        SELECT af.app_id,
                                                                               af.formation_id
                                                                        FROM apps_formations_id af) formation_apps ON b.app_id = formation_apps.id
WHERE event_def_id IS NOT NULL;

--

DROP VIEW IF EXISTS aspect_api_resources;

CREATE VIEW aspect_api_resources AS
SELECT aspects.id                       AS aspect_id,
       entries."ordId"                  AS ord_id,
       entries."minVersion"             AS min_version,
       formation_apps.formation_id as formation_id
FROM aspects JOIN (SELECT a1.id,
                          'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                   FROM applications a1
                   UNION ALL
                   SELECT af.app_id,
                          af.formation_id
                   FROM apps_formations_id af) formation_apps ON aspects.app_id = formation_apps.id, jsonb_to_recordset(aspects.api_resources) AS entries("ordId" TEXT, "minVersion" TEXT);

--

DROP VIEW IF EXISTS tenants_destinations;

CREATE VIEW tenants_destinations(tenant_id, formation_id, id, name, type, url, authentication, bundle_id, revision, sensitive_data) AS
SELECT DISTINCT dests.tenant_id,
                formation_apps.formation_id,
                dests.id,
                dests.name,
                dests.type,
                dests.url,
                dests.authentication,
                dests.bundle_id,
                dests.revision,
                '__sensitive_data__' || dests.name || '__sensitive_data__'
FROM destinations dests JOIN bundles b ON dests.bundle_id = b.id JOIN (SELECT a1.id,
                                                                              'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                       FROM applications a1
                                                                       UNION ALL
                                                                       SELECT af.app_id,
                                                                              af.formation_id
                                                                       FROM apps_formations_id af) formation_apps ON b.app_id = formation_apps.id;

--

DROP VIEW IF EXISTS api_product;

CREATE VIEW api_product AS
SELECT api_definitions.id     AS api_definition_id,
       p.id                   AS product_id,
       formation_apps.formation_id
FROM api_definitions, jsonb_array_elements_text(api_definitions.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id LEFT JOIN (SELECT a1.id, --left join so that we keep the global products the app_id of which will be null
                                                                        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                 FROM applications a1
                                                                 UNION ALL
                                                                 SELECT af.app_id,
                                                                        af.formation_id
                                                                 FROM apps_formations_id af) formation_apps ON p.app_id = formation_apps.id WHERE p.app_id = api_definitions.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS event_product;

CREATE VIEW event_product AS
SELECT event_api_definitions.id     AS event_definition_id,
       p.id                         AS product_id,
       formation_apps.formation_id
FROM event_api_definitions, jsonb_array_elements_text(event_api_definitions.part_of_products) AS elements
                          JOIN products p ON elements.value = p.ord_id LEFT JOIN (SELECT a1.id, --left join so that we keep the global products the app_id of which will be null
                                                                                         'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                                  FROM applications a1
                                                                                  UNION ALL
                                                                                  SELECT af.app_id,
                                                                                         af.formation_id
                                                                                  FROM apps_formations_id af) formation_apps ON p.app_id = formation_apps.id WHERE p.app_id = event_api_definitions.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS package_product;

CREATE VIEW package_product AS
SELECT pkgs.id          AS package_id,
       pr.id            AS product_id,
       formation_apps.formation_id
FROM packages pkgs, jsonb_array_elements_text(pkgs.part_of_products) AS elements
                        JOIN products pr ON elements.value = pr.ord_id LEFT JOIN (SELECT a1.id, -- left join so that we keep the global products the app_id of which will be null
                                                                                         'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                                  FROM applications a1
                                                                                  UNION ALL
                                                                                  SELECT af.app_id,
                                                                                         af.formation_id
                                                                                  FROM apps_formations_id af) formation_apps ON pr.app_id = formation_apps.id
WHERE pr.app_id = pkgs.app_id OR pr.app_id IS NULL;

--

DROP VIEW IF EXISTS entity_type_product;

CREATE VIEW entity_type_product AS
SELECT et.id           AS entity_type_id,
       p.id            AS product_id,
       formation_apps.formation_id
FROM entity_types et, jsonb_array_elements_text(et.part_of_products) AS elements
                          JOIN products p ON elements.value = p.ord_id LEFT JOIN (SELECT a1.id, -- left join so that we keep the global products the app_id of which will be null
                                                                                         'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS formation_id
                                                                                  FROM applications a1
                                                                                  UNION ALL
                                                                                  SELECT af.app_id,
                                                                                         af.formation_id
                                                                                  FROM apps_formations_id af) formation_apps ON p.app_id = formation_apps.id
WHERE p.app_id = et.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS tenants_entity_type_mappings;

CREATE OR REPLACE VIEW tenants_entity_type_mappings
            (tenant_id, formation_id, id, api_definition_id, event_definition_id, api_model_selectors, entity_type_targets)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                t_api_event_def.formation_id,
                etm.id,
                etm.api_definition_id,
                etm.event_definition_id,
                etm.api_model_selectors,
                etm.entity_type_targets
FROM entity_type_mappings etm
         JOIN (SELECT a.id,
                      a.tenant_id,
                      a.formation_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id,
                      e.formation_id
               FROM tenants_events e) t_api_event_def
              ON etm.api_definition_id = t_api_event_def.id OR etm.event_definition_id = t_api_event_def.id;

COMMIT;
