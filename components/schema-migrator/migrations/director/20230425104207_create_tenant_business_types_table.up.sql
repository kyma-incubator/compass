BEGIN;

CREATE TABLE tenant_business_types (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    code varchar(10) NOT NULL,
    name varchar(36) NOT NULL
);

ALTER TABLE tenant_business_types
    ADD CONSTRAINT tenant_business_types_code_name_unique UNIQUE (code, name);

ALTER TABLE applications ADD COLUMN tenant_business_type_id uuid;

ALTER TABLE applications
    ADD CONSTRAINT applications_tenant_business_type_id_fk
        FOREIGN KEY (tenant_business_type_id) REFERENCES tenant_business_types (id) ON DELETE SET NULL;

COMMIT;