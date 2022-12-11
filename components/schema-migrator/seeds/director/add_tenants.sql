INSERT INTO public.business_tenant_mappings (id, external_name, external_tenant, provider_name, status, type) VALUES
('3e64ebae-38b5-46a0-b1ed-9ccee153a0ae', 'default', '3e64ebae-38b5-46a0-b1ed-9ccee153a0ae', 'Compass', 'Active', 'account'),
('1eba80dd-8ff6-54ee-be4d-77944d17b10b', 'foo', '1eba80dd-8ff6-54ee-be4d-77944d17b10b', 'Compass', 'Active', 'account'),
('af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e', 'bar', 'af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e', 'Compass', 'Active', 'account');

INSERT INTO formation_templates(id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, tenant_id)
VALUES (uuid_generate_v4(), 'Side-by-side extensibility with Kyma',
        '["SAP Cloud for Customer", "SAP Commerce Cloud", "SAP Field Service Management", "SAP Marketing Cloud"]',
        '["kyma"]',
        'SAP BTP Kyma',
        'ENVIRONMENT_INSTANCE',
        '3e64ebae-38b5-46a0-b1ed-9ccee153a0ae'
       );


