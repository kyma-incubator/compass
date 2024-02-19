BEGIN;

CREATE INDEX  operation_on_op_type ON operation (op_type);
CREATE INDEX  operation_on_priority ON operation (priority DESC);
CREATE INDEX  operation_on_status ON operation (status);
CREATE INDEX  operation_on_updated_at ON operation (updated_at);

COMMIT;
