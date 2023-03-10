BEGIN;

UPDATE formation_constraints
SET name = 'SubaccountInAtMostOneFormationOfGivenType'
WHERE name = 'SubaccountInAtMostOneEventMeshFormation';

INSERT INTO formation_constraints (id, name, constraint_type, target_operation, operator, resource_type,
                                   resource_subtype, input_template, constraint_scope)
VALUES (uuid_generate_v4(), 'SystemInAtMostOneFormationOfGivenType', 'PRE', 'ASSIGN_FORMATION', 'IsNotAssignedToAnyFormationOfType', 'APPLICATION',
        'ANY', '{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}","exceptSystemTypes": ["SAP S/4HANA Cloud"]}', 'FORMATION_TYPE');

COMMIT;