BEGIN;

ALTER TABLE operation
    DROP COLUMN error_severity;

DROP TYPE IF EXISTS operation_error_severity;

COMMIT;
