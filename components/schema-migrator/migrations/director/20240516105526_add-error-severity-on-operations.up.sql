BEGIN;

DROP VIEW IF EXISTS scheduled_operations;

CREATE TYPE operation_error_severity AS ENUM (
    'ERROR',
    'WARNING',
    'INFO',
    'NONE'
);

ALTER TABLE operation
    ADD COLUMN error_severity operation_error_severity;

CREATE VIEW scheduled_operations AS
    SELECT id, op_type, status, data, error, error_severity, priority, created_at, updated_at
    FROM operation
    WHERE status = 'SCHEDULED'
    ORDER BY priority DESC, updated_at;

COMMIT;
