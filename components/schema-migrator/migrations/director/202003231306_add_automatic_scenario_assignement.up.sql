CREATE TABLE automatic_scenario_assignements (
    scenario VARCHAR(128),
    tenant_id UUID NOT NULL REFERENCES business_tenant_mappings(id),
    key VARCHAR(256) NOT NULL,
    value JSONB);


CREATE INDEX ON automatic_scenario_assignements (tenant_id);

ALTER TABLE automatic_scenario_assignements
ADD CONSTRAINT automatic_scenario_assignements_pk PRIMARY KEY (tenant_id, scenario);