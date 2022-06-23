BEGIN;

INSERT INTO labels(id, tenant_id, app_id, runtime_id, key, value, runtime_context_id, version, app_template_id)
SELECT uuid_generate_v4() as id, NULL::uuid as tenant_id, NULL::uuid as app_id, t.runtime_id, 'runtimeType' as key, '"kyma"'::jsonb as value, NULL::uuid as runtime_context_id, 0 as version, NULL::uuid as app_template_id
FROM (SELECT DISTINCT runtime_id FROM labels WHERE runtime_id IS NOT NULL
      EXCEPT
      SELECT runtime_id FROM labels where key = 'xsappname') as t;

COMMIT;
