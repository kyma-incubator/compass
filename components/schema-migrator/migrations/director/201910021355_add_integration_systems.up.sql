CREATE TABLE integration_systems
(
    id   uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name varchar(255) NOT NULL
);

CREATE UNIQUE INDEX ON integration_systems (name);

ALTER TABLE system_auths
    ADD CONSTRAINT system_auths_integration_systems_id_fk
        FOREIGN KEY (integration_system_id) REFERENCES integration_systems (id) ON DELETE CASCADE;
