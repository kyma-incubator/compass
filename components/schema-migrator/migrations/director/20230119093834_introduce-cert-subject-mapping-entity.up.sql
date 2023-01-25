BEGIN;

CREATE TABLE cert_subject_mapping (
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    subject VARCHAR(256) NOT NULL,
    CONSTRAINT cert_mapping_subject_unique UNIQUE (subject),
    consumer_type VARCHAR(256) NOT NULL,
    internal_consumer_id UUID CHECK (internal_consumer_id <> '00000000-0000-0000-0000-000000000000'),
    tenant_access_levels jsonb NOT NULL
);

COMMIT;
