BEGIN;

-- Drop views
DROP VIEW IF EXISTS tags_capabilities;
DROP VIEW IF EXISTS links_capabilities;
DROP VIEW IF EXISTS ord_labels_capabilities;
DROP VIEW IF EXISTS correlation_ids_capabilities;
DROP VIEW IF EXISTS ord_documentation_labels_capabilities;

DROP VIEW IF EXISTS capability_definitions;

DROP VIEW IF EXISTS api_specifications_tenants;
DROP VIEW IF EXISTS event_specifications_tenants;
DROP VIEW IF EXISTS capability_specifications_tenants;
DROP VIEW IF EXISTS capabilities_tenants;

DROP VIEW IF EXISTS tenants_specifications;
DROP VIEW IF EXISTS tenants_capabilities;

-- Alter table specifications - remove columns capability_def_id, capability_spec_type and capability_spec_format
ALTER TABLE specifications
    DROP CONSTRAINT specifications_capability_id_fkey,
    DROP COLUMN capability_def_id,
    DROP COLUMN capability_spec_type,
    DROP COLUMN capability_spec_format;

-- Drop types associated with capability specification
DROP TYPE capability_spec_type;
DROP TYPE capability_spec_format;

-- Drop table capability
DROP TABLE IF EXISTS capabilities;
-- Drop type associated with capability
DROP TYPE capability_type;

-- Recreate views api_specifications_tenants and event_specifications_tenants
CREATE OR REPLACE VIEW api_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                             INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ad.app_id);


CREATE OR REPLACE VIEW event_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                            INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                            INNER JOIN tenant_applications ta on ta.id = ead.app_id);


-- Recreate view tenants_specifications
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
