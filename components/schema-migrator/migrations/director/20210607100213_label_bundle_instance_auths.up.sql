BEGIN;

DO $$
    DECLARE
        rec RECORD;
        app_scenarios JSONB;
        runtime_scenarios JSONB;
        common_scenarios JSONB;
    BEGIN

        CREATE OR REPLACE VIEW bundle_instance_auths_scenarios AS SELECT bundle_instance_auths.id, bundles.tenant_id, app_id, runtime_id
                                                                  FROM bundle_instance_auths INNER JOIN bundles ON bundle_instance_auths.bundle_id=bundles.id;

        FOR rec IN SELECT "id", tenant_id, app_id, runtime_id FROM bundle_instance_auths_scenarios LOOP
                SELECT "value" INTO app_scenarios FROM labels WHERE "key"='scenarios' AND app_id=rec.app_id;
                SELECT "value" INTO runtime_scenarios FROM labels WHERE "key"='scenarios' AND runtime_id=rec.runtime_id;

                SELECT jsonb_agg(scenarios.jsonb_array_elements) INTO common_scenarios FROM (SELECT (jsonb_array_elements(app_scenarios))
                                                                                             INTERSECT
                                                                                             SELECT (jsonb_array_elements(runtime_scenarios)))scenarios;

                RAISE INFO 'Scenarios for bundle instance auth with id % : %', rec.id, common_scenarios;

                IF NOT EXISTS (SELECT * FROM labels WHERE bundle_instance_auth_id=rec.id) THEN
                    RAISE INFO 'Labeling bundle instance auth with id % with scenarios', rec.id;
                    INSERT INTO labels ("id",tenant_id,"key","value",bundle_instance_auth_id)
                    VALUES (uuid_generate_v4(), rec.tenant_id, 'scenarios', common_scenarios, rec.id);
                END IF;

            END LOOP;

        DROP VIEW bundle_instance_auths_scenarios;

    EXCEPTION WHEN OTHERS THEN ROLLBACK;

    END; $$;

ALTER TABLE labels ADD COLUMN bundle_instance_auth_id uuid REFERENCES bundle_instance_auths(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS labels_tenant_id_key_coalesce_coalesce1_coalesce2_idx;
CREATE UNIQUE INDEX labels_default_values ON labels (tenant_id, key, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_context_id, '00000000-0000-0000-0000-000000000000'), coalesce(bundle_instance_auth_id, '00000000-0000-0000-0000-000000000000'));

ALTER TABLE labels
DROP CONSTRAINT valid_refs;

ALTER TABLE labels
    ADD CONSTRAINT valid_refs
        CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL OR runtime_context_id IS NOT NULL OR bundle_instance_auth_id IS NOT NULL);

CREATE VIEW bundle_instance_auths_scenarios_labels AS
SELECT labels.key, labels.value, labels.tenant_id, labels.id as label_id, bundles.id as bundle_id, bundles.app_id, bundle_instance_auths.runtime_id, bundle_instance_auths.id as bundle_instance_auth_id, bundle_instance_auths.status_condition
FROM labels
         INNER JOIN bundle_instance_auths ON labels.bundle_instance_auth_id = bundle_instance_auths.id
         INNER JOIN bundles ON bundles.id = bundle_instance_auths.bundle_id
WHERE labels.key='scenarios' AND bundle_instance_auths.status_condition <> 'UNUSED';

COMMIT;
