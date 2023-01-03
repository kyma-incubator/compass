package model

type FormationConstraintType string

const (
	PreOperation  FormationConstraintType = "pre"
	PostOperation FormationConstraintType = "post"
)

type TargetOperation string

const (
	AssignFormationOperation      TargetOperation = "assignFormation"
	UnassignFormationOperation    TargetOperation = "unassignFormation"
	CreateFormationOperation      TargetOperation = "creatFormation"
	DeleteFormationOperation      TargetOperation = "deleteFormation"
	GenerateNotificationOperation TargetOperation = "generateNotification"
)

type ResourceType string

const (
	ApplicationResourceType ResourceType = "application"
	RuntimeResourceType     ResourceType = "runtime"
	TenantResourceType      ResourceType = "tenant"
	FormationResourceType   ResourceType = "formation"
)

type OperatorScopeType string

const (
	TenantScope    OperatorScopeType = "tenant"
	FormationScope OperatorScopeType = "formation"
	GlobalScope    OperatorScopeType = "global"
)

type FormationConstraintScope string

const (
	GlobalFormationConstraintScope        FormationConstraintScope = "global"
	FormationTypeFormationConstraintScope FormationConstraintScope = "formation_type"
)

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
