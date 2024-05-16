BEGIN;

DROP VIEW IF EXISTS scheduled_operations;

ALTER TABLE operation DROP COLUMN error_severity;

DROP TYPE IF EXISTS operation_error_severity;

CREATE VIEW scheduled_operations AS
SELECT id, op_type, status, data, error, priority, created_at, updated_at
    FROM operation
    WHERE status = 'SCHEDULED'
    ORDER BY priority DESC, updated_at;

COMMIT;
