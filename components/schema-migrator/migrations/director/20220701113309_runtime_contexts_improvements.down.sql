BEGIN;

CREATE OR REPLACE VIEW runtime_contexts_labels_tenants AS
SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
    INNER JOIN runtime_contexts rc ON l.runtime_context_id = rc.id
    INNER JOIN tenant_runtimes tr ON rc.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id);


-- Aggregated labels view
CREATE OR REPLACE VIEW labels_tenants AS
(SELECT l.id, ta.tenant_id, ta.owner FROM labels AS l
    INNER JOIN tenant_applications ta
        ON l.app_id = ta.id AND (l.tenant_id IS NULL OR l.tenant_id = ta.tenant_id))
UNION ALL
(SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
    INNER JOIN tenant_runtimes tr
        ON l.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id))
UNION ALL
(SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
    INNER JOIN runtime_contexts rc ON l.runtime_context_id = rc.id
    INNER JOIN tenant_runtimes tr ON rc.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id));


DROP INDEX labels_runtime_context_id;
DROP INDEX tenant_runtime_contexts_rtm_context_id;
DROP INDEX tenant_runtime_contexts_tenant_id;

COMMIT;
