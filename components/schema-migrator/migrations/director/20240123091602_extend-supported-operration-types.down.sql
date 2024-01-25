BEGIN;

ALTER TABLE operation
    DROP CONSTRAINT operation_op_type_check;

ALTER TABLE operation
    ADD CONSTRAINT operation_op_type_check CHECK (op_type IN ('ORD_AGGREGATION'));

COMMIT;


