package model

type FormationConstraintType string

const (
	PreOperation  FormationConstraintType = "PRE"
	PostOperation FormationConstraintType = "POST"
)

type TargetOperation string

const (
	AssignFormationOperation      TargetOperation = "ASSIGN_FORMATION"
	UnassignFormationOperation    TargetOperation = "UNASSIGNED_FORMATION"
	CreateFormationOperation      TargetOperation = "CREATE_FORMATION"
	DeleteFormationOperation      TargetOperation = "DELETE_FORMATION"
	GenerateNotificationOperation TargetOperation = "GENERATE_NOTIFICATION"
)

type ResourceType string

const (
	ApplicationResourceType    ResourceType = "APPLICATION"
	RuntimeResourceType        ResourceType = "RUNTIME"
	RuntimeContextResourceType ResourceType = "RUNTIME_CONTEXT"
	TenantResourceType         ResourceType = "TENANT"
	FormationResourceType      ResourceType = "FORMATION"
)

type OperatorScopeType string

const (
	TenantScope    OperatorScopeType = "TENANT"
	FormationScope OperatorScopeType = "FORMATION"
	GlobalScope    OperatorScopeType = "GLOBAL"
)

type FormationConstraintScope string

const (
	GlobalFormationConstraintScope        FormationConstraintScope = "GLOBAL"
	FormationTypeFormationConstraintScope FormationConstraintScope = "FORMATION_TYPE"
)

type FormationConstraintInput struct {
	Name                string
	ConstraintType      FormationConstraintType
	TargetOperation     TargetOperation
	Operator            string
	ResourceType        ResourceType
	ResourceSubtype     string
	OperatorScope       OperatorScopeType
	InputTemplate       string
	ConstraintScope     FormationConstraintScope
	FormationTemplateID string
}

type FormationConstraint struct {
	ID              string
	Name            string
	ConstraintType  FormationConstraintType
	TargetOperation TargetOperation
	Operator        string
	ResourceType    ResourceType
	ResourceSubtype string
	OperatorScope   OperatorScopeType
	InputTemplate   string
	ConstraintScope FormationConstraintScope
}
