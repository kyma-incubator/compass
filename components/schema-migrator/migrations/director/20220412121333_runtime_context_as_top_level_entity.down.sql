BEGIN;

CREATE
OR REPLACE VIEW runtime_contexts_tenants AS
SELECT rtc.*, tr.tenant_id, tr.owner
FROM runtime_contexts AS rtc
         INNER JOIN tenant_runtimes tr ON rtc.runtime_id = tr.id;

DROP TABLE tenant_runtime_contexts;

COMMIT;
