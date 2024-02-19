BEGIN;

DROP INDEX IF EXISTS operation_on_updated_at;
DROP INDEX IF EXISTS operation_on_status;
DROP INDEX IF EXISTS operation_on_priority;
DROP INDEX IF EXISTS operation_on_op_type;

COMMIT;
