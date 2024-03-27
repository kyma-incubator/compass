BEGIN;

DROP VIEW IF EXISTS tenants_api_bundle_reference;
DROP VIEW IF EXISTS tenants_event_bundle_reference;

--

DROP VIEW IF EXISTS aspect_api_resources;

CREATE VIEW aspect_api_resources AS
SELECT id                               AS aspect_id,
       entries."ordId"                  AS ord_id,
       entries."minVersion"             AS min_version
FROM aspects,
     jsonb_to_recordset(aspects.api_resources) AS entries("ordId" TEXT, "minVersion" TEXT); -- if the JSON key has a capital letter then it should be defined with "" in the AS clause

--

DROP VIEW IF EXISTS tenants_destinations;

CREATE VIEW tenants_destinations(tenant_id, id, name, type, url, authentication, bundle_id, revision, sensitive_data) AS
SELECT DISTINCT dests.tenant_id,
                dests.id,
                dests.name,
                dests.type,
                dests.url,
                dests.authentication,
                dests.bundle_id,
                dests.revision,
                '__sensitive_data__' || dests.name || '__sensitive_data__'
FROM destinations dests;

--

DROP VIEW IF EXISTS api_product;

CREATE VIEW api_product AS
SELECT api_definitions.id     AS api_definition_id,
       p.id                   AS product_id
FROM api_definitions,
     jsonb_array_elements_text(api_definitions.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = api_definitions.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS event_product;

CREATE VIEW event_product AS
SELECT event_api_definitions.id     AS event_definition_id,
       p.id                         AS product_id
FROM event_api_definitions,
     jsonb_array_elements_text(event_api_definitions.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = event_api_definitions.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS package_product;

CREATE VIEW package_product AS
SELECT packages.id     AS package_id,
       p.id            AS product_id
FROM packages,
     jsonb_array_elements_text(packages.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = packages.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS entity_type_product;

CREATE VIEW entity_type_product AS
SELECT entity_types.id  AS entity_type_id,
       p.id             AS product_id
FROM entity_types,
     jsonb_array_elements_text(entity_types.part_of_products) AS elements
         JOIN products p ON elements.value = p.ord_id WHERE p.app_id = entity_types.app_id OR p.app_id IS NULL;

--

DROP VIEW IF EXISTS tenants_entity_type_mappings;

CREATE VIEW tenants_entity_type_mappings
            (tenant_id, id, api_definition_id, event_definition_id, api_model_SELECTors, entity_type_targets)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                etm.id,
                etm.api_definition_id,
                etm.event_definition_id,
                etm.api_model_SELECTors,
                etm.entity_type_targets
FROM entity_type_mappings etm
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e) t_api_event_def
              ON etm.api_definition_id = t_api_event_def.id OR etm.event_definition_id = t_api_event_def.id;

COMMIT;
