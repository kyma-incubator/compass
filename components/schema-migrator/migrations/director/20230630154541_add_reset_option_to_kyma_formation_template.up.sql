BEGIN;

INSERT INTO formation_constraints(id, name, constraint_type, target_operation, operator, resource_type, resource_subtype, input_template, constraint_scope)
VALUES(uuid_generate_v4(), 'DoNotSendNotificationsToKymaAdapterInAFormationTypeExceptForServiceCloudVersion2', 'PRE', 'GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION', 'DoNotGenerateFormationAssignmentNotification', 'RUNTIME', 'kyma', '{"resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","source_resource_type": "APPLICATION","source_resource_id": "{{.Application.ID}}","tenant": "{{.TenantID}}","formation_template_id":"{{.FormationTemplateID}}","except_subtypes": ["SAP Service Cloud Version 2", "App1"]}', 'GLOBAL');

UPDATE formation_templates SET supports_reset = TRUE WHERE name = 'Side-by-Side Extensibility with Kyma';

COMMIT;
