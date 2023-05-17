BEGIN;

CREATE TYPE operation_status AS ENUM (
    'SCHEDULED',
    'IN_PROGRESS',
    'COMPLETED',
    'FAILED'
);

CREATE TABLE IF NOT EXISTS operation
(
    id          UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    op_type     VARCHAR(256) NOT NULL CHECK ( op_type in ('ORD_AGGREGATION') ),
    status      operation_status NOT NULL,
    data        JSONB,
    error       JSONB,
    priority    INTEGER DEFAULT 1,
    created_at  TIMESTAMP,
    finished_at TIMESTAMP
);

COMMIT;
