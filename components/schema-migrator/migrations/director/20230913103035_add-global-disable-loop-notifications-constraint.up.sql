BEGIN;

INSERT INTO formation_constraints(id, name, description, constraint_type, target_operation, operator, resource_type, resource_subtype, input_template, constraint_scope, created_at)
VALUES
    (uuid_generate_v4(),'DoNotGenerateFormationAssignmentNotificationForLoopsGlobalApplication', 'Skip generating notifications for loops for application that is assigned' , 'PRE', 'GENERATE_FORMATION_ASSIGNMENT_NOTIFICATION', 'DoNotGenerateFormationAssignmentNotificationForLoops', 'APPLICATION', 'ANY', '{"resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","source_resource_type": "APPLICATION","source_resource_id":"{{ if .SourceApplication }}{{.SourceApplication.ID}}{{else}}{{.Assignment.Source}}{{end}}","tenant": "{{.TenantID}}"}', 'GLOBAL', CURRENT_TIMESTAMP),
COMMIT;
