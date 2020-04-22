CREATE TABLE IF NOT EXISTS lms_tenants (
    id varchar(255) PRIMARY KEY,
    name varchar(255) NOT NULL,
    region varchar(12) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    unique (name, region)
);
