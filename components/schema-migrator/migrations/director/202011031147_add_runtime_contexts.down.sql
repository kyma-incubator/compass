BEGIN;

ALTER TABLE labels
    DROP CONSTRAINT runtime_context_id_fk;

ALTER TABLE labels
    DROP COLUMN runtime_context_id;

ALTER TABLE labels
    DROP CONSTRAINT valid_refs;

ALTER TABLE labels
    ADD CONSTRAINT valid_refs
        CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL);

DROP INDEX IF EXISTS labels_tenant_id_key_coalesce_coalesce1_idx;

CREATE UNIQUE INDEX ON labels (tenant_id, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'));

DROP TABLE runtime_contexts;

COMMIT;
