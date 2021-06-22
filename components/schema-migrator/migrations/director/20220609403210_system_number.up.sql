BEGIN;

ALTER TABLE applications
    ADD COLUMN system_number VARCHAR(256) NULL;

ALTER TABLE applications DROP CONSTRAINT application_tenant_id_name_unique;

ALTER TABLE applications
    ADD CONSTRAINT application_tenant_id_name_unique UNIQUE (tenant_id, name, system_number);

ALTER TYPE application_status_condition ADD VALUE 'MANAGED';

alter table applications alter column "name" type varchar(256);

COMMIT;
