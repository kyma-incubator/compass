BEGIN;

DROP VIEW IF EXISTS tags_capabilities;
DROP VIEW IF EXISTS links_capabilities;
DROP VIEW IF EXISTS ord_labels_capabilities;
DROP VIEW IF EXISTS correlation_ids_capabilities;
DROP VIEW IF EXISTS ord_documentation_labels_capabilities;
DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_capabilities;
DROP VIEW IF EXISTS capability_definitions;

ALTER TABLE specifications
    DROP CONSTRAINT specifications_capability_id_fkey,
    DROP COLUMN capability_def_id,
    DROP COLUMN capability_spec_type,
    DROP COLUMN capability_spec_format;

DROP TYPE capability_spec_type;
DROP TYPE capability_spec_format;

DROP TABLE IF EXISTS capabilities;
DROP TYPE capability_type;


CREATE OR REPLACE VIEW tenants_specifications
            (tenant_id, id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format,
             event_spec_type, custom_type, created_at)
AS
SELECT DISTINCT t_api_event_def.tenant_id,
                spec.id,
                spec.api_def_id,
                spec.event_def_id,
                spec.spec_data,
                spec.api_spec_format,
                spec.api_spec_type,
                spec.event_spec_format,
                spec.event_spec_type,
                spec.custom_type,
                spec.created_at
FROM specifications spec
         JOIN (SELECT a.id,
                      a.tenant_id
               FROM tenants_apis a
               UNION ALL
               SELECT e.id,
                      e.tenant_id
               FROM tenants_events e) t_api_event_def
              ON spec.api_def_id = t_api_event_def.id OR spec.event_def_id = t_api_event_def.id;


COMMIT;
