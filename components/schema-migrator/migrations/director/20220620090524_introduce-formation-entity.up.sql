BEGIN;

CREATE TABLE formations (
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    tenant_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE,
    formation_template_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    FOREIGN KEY (formation_template_id) REFERENCES formation_templates(id) ON DELETE CASCADE,
    name VARCHAR(256) NOT NULL
);

INSERT INTO formations(id, tenant_id, formation_template_id, name)
SELECT uuid_generate_v4(), ld.tenant_id, (SELECT ft.id FROM formation_templates ft WHERE ft.name = 'Side-by-side extensibility with Kyma') as formation_template_id, jsonb_array_elements_text(schema->'items'->'enum') as name
FROM label_definitions ld
WHERE ld.key = 'scenarios';

COMMIT;
