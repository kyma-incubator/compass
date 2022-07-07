BEGIN;

CREATE OR REPLACE VIEW runtime_contexts_labels_tenants AS
SELECT l.id, trc.tenant_id, trc.owner FROM labels AS l
    INNER JOIN tenant_runtime_contexts trc
        ON l.runtime_context_id = trc.id AND (l.tenant_id IS NULL OR l.tenant_id = trc.tenant_id);


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
(SELECT l.id, trc.tenant_id, trc.owner FROM labels AS l
    INNER JOIN tenant_runtime_contexts trc
        ON l.runtime_context_id = trc.id AND (l.tenant_id IS NULL OR l.tenant_id = trc.tenant_id));


CREATE INDEX labels_runtime_context_id
    ON labels (runtime_context_id)
    WHERE (labels.runtime_context_id IS NOT NULL);


CREATE INDEX tenant_runtime_contexts_rtm_context_id
    ON tenant_runtime_contexts (id);


CREATE INDEX tenant_runtime_contexts_tenant_id
    ON tenant_runtime_contexts (tenant_id);

COMMIT;
