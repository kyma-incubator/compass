BEGIN;

CREATE TABLE formation_constraints
(
    id               UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name             VARCHAR(256) NOT NULL,
    constraint_type  VARCHAR(256) NOT NULL CHECK ( constraint_type in ('PRE', 'POST') ),
    target_operation VARCHAR(256) NOT NULL CHECK ( target_operation in ('ASSIGN_FORMATION','UNASSIGN_FORMATION','CREATE_FORMATION','DELETE_FORMATION','GENERATE_NOTIFICATION') ),
    operator         VARCHAR(256) NOT NULL,
    resource_type    VARCHAR(256) NOT NULL CHECK ( resource_type in ('APPLICATION', 'RUNTIME', 'RUNTIME_CONTEXT','TENANT','FORMATION') ),
    resource_subtype VARCHAR(256) NOT NULL,
    input_template   TEXT,
    constraint_scope VARCHAR(256) NOT NULL CHECK ( constraint_scope in ('GLOBAL','FORMATION_TYPE') )
);

CREATE TABLE formation_template_constraint_references
(
    formation_template_id    UUID NOT NULL    CHECK (formation_template_id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (formation_template_id) REFERENCES formation_templates(id) ON DELETE CASCADE,
    formation_constraint_id    UUID NOT NULL    CHECK (formation_constraint_id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (formation_constraint_id) REFERENCES formation_constraints(id) ON DELETE CASCADE
);

CREATE INDEX formation_template_id ON formation_template_constraint_references (formation_template_id);
CREATE INDEX formation_constraint_id ON formation_template_constraint_references (formation_template_id);

-- Create constraint that does not allow a Subaccount to be added to multiple EM formations
insert into formation_constraints (id, name, constraint_type, target_operation, operator, resource_type,
                                   resource_subtype, input_template, constraint_scope)
values (uuid_generate_v4(), 'SubaccountInAtMostOneEventMeshFormation', 'PRE', 'ASSIGN_FORMATION', 'participatesInFormationsOfType', 'TENANT',
        'subaccount', '{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}"}', 'FORMATION_TYPE');

COMMIT;
