BEGIN;

INSERT INTO tenant_applications
    (SELECT tp.parent_id, ta.id, ta.owner, ta.tenant_id
     FROM tenant_applications ta
              JOIN tenant_parents tp ON ta.tenant_id = tp.tenant_id
     WHERE NOT EXISTS(SELECT 1
                      FROM tenant_applications ti
                      WHERE ta.owner = ti.owner
                        AND ta.id = ti.id
                        AND ta.tenant_id = ti.source
                        AND ta.source = ti.tenant_id)) ON CONFLICT (tenant_id,id,source) DO NOTHING;

CREATE INDEX business_tenant_mappings_type_idx ON business_tenant_mappings (type);

CREATE INDEX labels_key_idx ON labels (key);

CREATE INDEX formation_assignments_source_idx ON formation_assignments (source);

CREATE INDEX formation_assignments_target_idx ON formation_assignments (target);

COMMIT;
