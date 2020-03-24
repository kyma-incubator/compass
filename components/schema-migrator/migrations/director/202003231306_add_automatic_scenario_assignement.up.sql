CREATE TABLE automatic_scenario_assignments (
    scenario VARCHAR(128),
    tenant_id UUID NOT NULL,
    key VARCHAR(256) NOT NULL,
    value VARCHAR(256));


CREATE INDEX ON automatic_scenario_assignments (tenant_id);

ALTER TABLE automatic_scenario_assignments
    ADD CONSTRAINT automatic_scenario_assignments_pk
    PRIMARY KEY (tenant_id, scenario);

ALTER TABLE automatic_scenario_assignments
    ADD CONSTRAINT automatic_scenario_assignments_tenant_constraint
    FOREIGN KEY (tenant_id)
    REFERENCES business_tenant_mappings(id)
    ON DELETE CASCADE;