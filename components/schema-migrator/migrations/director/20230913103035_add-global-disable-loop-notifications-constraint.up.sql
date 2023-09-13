BEGIN;

INSERT INTO formation_constraints(id, name, constraint_type, target_operation, operator, resource_type, resource_subtype, input_template, constraint_scope)
VALUES
    (uuid_generate_v4(),'DoNotGenerateFormationAssignmentNotificationForLoopsGlobalApplication', 'PRE', 'GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION', 'DoNotGenerateFormationAssignmentNotificationForLoops', 'APPLICATION', 'ANY', '{"resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","source_resource_type": "APPLICATION","source_resource_id":"{{ if .SourceApplication }}{{.SourceApplication.ID}}{{else}}{{.Assignment.Source}}{{end}}","tenant": "{{.TenantID}}"}', 'GLOBAL'),
    (uuid_generate_v4(),'DoNotGenerateFormationAssignmentNotificationForLoopsGlobalRuntime', 'PRE', 'GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION', 'DoNotGenerateFormationAssignmentNotificationForLoops', 'RUNTIME', 'ANY', '{"resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","source_resource_type": "APPLICATION","source_resource_id": "{{.Assignment.Source}}", "tenant": "{{.TenantID}}"}', 'GLOBAL'),
    (uuid_generate_v4(),'DoNotGenerateFormationAssignmentNotificationForLoopsGlobalRuntimeContext', 'PRE', 'GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION', 'DoNotGenerateFormationAssignmentNotificationForLoops', 'RUNTIME_CONTEXT', 'ANY', '{"resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","source_resource_type": "RUNTIME_CONTEXT","source_resource_id": "{{.Assignment.Source}}", "tenant": "{{.TenantID}}"}', 'GLOBAL');

COMMIT;
