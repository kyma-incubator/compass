CREATE TYPE runtime_status_condition AS ENUM ('INITIAL', 'READY', 'FAILED');

CREATE TABLE runtime
(
    id uuid NOT NULL CONSTRAINT runtime_pk PRIMARY KEY,
    tenant_id uuid NOT NULL,
    name varchar(256) NOT NULL,
    description text,
    status_condition runtime_status_condition DEFAULT 'INITIAL'::runtime_status_condition NOT NULL,
    status_timestamp timestamp NOT NULL,
    auth json
);

CREATE UNIQUE INDEX runtime_id_uindex ON runtime (id);
CREATE UNIQUE INDEX runtime_id_name_uindex ON runtime (tenant_id, name);
