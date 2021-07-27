BEGIN;

ALTER TABLE labels
    ADD CONSTRAINT valid_refs
        CHECK (app_id IS NOT NULL OR runtime_id IS NOT NULL OR labels.runtime_context_id IS NOT NULL);

COMMIT;
