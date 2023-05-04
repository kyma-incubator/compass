BEGIN;

CREATE TABLE systems_sync_timestamps
(
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id UUID NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    last_sync_timestamp TIMESTAMP NOT NULL
);

ALTER TABLE systems_sync_timestamps
    ADD CONSTRAINT systems_sync_timestamps_tenant_id_product_id UNIQUE (tenant_id, product_id);

COMMIT;