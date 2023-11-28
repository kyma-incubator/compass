BEGIN;

-- Create many to many tenant_parents table
CREATE TABLE tenant_parents
(
    tenant_id   uuid NOT NULL,
    parent_id   uuid NOT NULL,

    CONSTRAINT tenant_parents_tenant_id_fkey  FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    CONSTRAINT tenant_parents_parent_id_fkey  FOREIGN KEY (parent_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    PRIMARY KEY (tenant_id, parent_id)
);

-- Create indexes for tenant_parents table
CREATE INDEX tenant_parents_tenant_id ON tenant_parents(tenant_id);
CREATE INDEX tenant_parents_parent_id ON tenant_parents(parent_id);


-- Populate 'tenant_parents' table with data from 'business_tenant_mappings'
INSERT INTO tenant_parents (tenant_id, parent_id)
SELECT id, parent
FROM business_tenant_mappings
WHERE parent IS NOT NULL;

-- Drop 'parent' column from 'business_tenant_mappings'
ALTER TABLE business_tenant_mappings
    DROP COLUMN parent;


-- TODO make source column not null
-- Add source column to tenant_applications table
ALTER TABLE tenant_applications
    ADD COLUMN source uuid;
ALTER TABLE tenant_applications
    ADD CONSTRAINT tenant_applications_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_applications
    ADD PRIMARY KEY (tenant_id, id, source);

-- Add source column to tenant_runtimes table
ALTER TABLE tenant_runtimes
    ADD COLUMN source uuid;
ALTER TABLE tenant_runtimes
    ADD CONSTRAINT tenant_runtimes_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_runtimes
    ADD PRIMARY KEY (tenant_id, id, source);

-- Add source column to tenant_runtime_contexts table
ALTER TABLE tenant_runtime_contexts
    ADD COLUMN source uuid;
ALTER TABLE tenant_runtime_contexts
    ADD CONSTRAINT tenant_runtime_contexts_source_fk
        FOREIGN KEY (source) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE;
ALTER TABLE tenant_runtime_contexts
    ADD PRIMARY KEY (tenant_id, id, source);

COMMIT;
