BEGIN;

ALTER TABLE labels drop CONSTRAINT valid_refs;

ALTER TABLE labels ADD CONSTRAINT valid_refs CHECK (num_nonnulls(labels.app_id, labels.runtime_id, labels.runtime_context_id) = 1 OR
                                                    (num_nonnulls(labels.app_id, labels.runtime_id, labels.runtime_context_id) = 0));

COMMIT;