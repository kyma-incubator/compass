BEGIN;

CREATE TABLE IF NOT EXISTS operation
(
    id          UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    op_type     VARCHAR(256) NOT NULL CHECK ( op_type in ('ORD_AGGREGATION') ),
    status      VARCHAR(256) NOT NULL CHECK ( status in ('SCHEDULED','IN_PROGRESS','COMPLETED','FAILED') ),
    data        JSONB,
    error       JSONB,
    priority    INTEGER DEFAULT 1,
    created_at  TIMESTAMP,
    finished_at TIMESTAMP
);

COMMIT;
