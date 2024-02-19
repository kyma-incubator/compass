BEGIN;

CREATE INDEX  operations_on_op_type ON operations (op_type);
CREATE INDEX  operations_on_priority ON operations (priority DESC);
CREATE INDEX  operations_on_status ON operations (status);
CREATE INDEX  operations_on_updated_at ON operations (updated_at);

COMMIT;
