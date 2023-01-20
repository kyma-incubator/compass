BEGIN;

CREATE TABLE formation_constraints
(
    id               UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name             VARCHAR(256) NOT NULL,
    constraint_type  VARCHAR(256) NOT NULL,
    target_operation VARCHAR(256) NOT NULL,
    operator         VARCHAR(256) NOT NULL,
    resource_type    VARCHAR(256) NOT NULL,
    resource_subtype VARCHAR(256) NOT NULL,
    operator_scope   VARCHAR(256) NOT NULL,
    input_template   TEXT,
    constraint_scope VARCHAR(256) NOT NULL
);

CREATE TABLE formation_template_constraint_references
(
    formation_template_id    UUID NOT NULL    CHECK (formation_template_id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (formation_template) REFERENCES formation_templates(id) ON DELETE CASCADE,
    formation_constraint_id    UUID NOT NULL    CHECK (formation_constraint_id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (formation_constraint) REFERENCES formation_constraints(id) ON DELETE CASCADE
);

CREATE INDEX formation_template_id ON formation_template_constraint_references (formation_template_id);
CREATE INDEX formation_constraint_id ON formation_template_constraint_references (formation_template_id);

COMMIT;
