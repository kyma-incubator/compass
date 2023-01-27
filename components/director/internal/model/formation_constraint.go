package model

// FormationConstraintType represents the constraint type. It is part of the Join point location along with TargetOperation
type FormationConstraintType string

const (
	// PreOperation denotes the constraint should be enforced before the operation execution
	PreOperation FormationConstraintType = "PRE"
	// PostOperation denotes the constraint should be enforced after the operation execution
	PostOperation FormationConstraintType = "POST"
)

// TargetOperation represents the operation to which the constraint is applicable
type TargetOperation string

const (
	// AssignFormationOperation represents the assign formation operation
	AssignFormationOperation TargetOperation = "ASSIGN_FORMATION"
	// UnassignFormationOperation represents the unassign formation operation
	UnassignFormationOperation TargetOperation = "UNASSIGN_FORMATION"
	// CreateFormationOperation represents the create formation operation
	CreateFormationOperation TargetOperation = "CREATE_FORMATION"
	// DeleteFormationOperation represents the delete formation operation
	DeleteFormationOperation TargetOperation = "DELETE_FORMATION"
	// GenerateNotificationOperation represents the generate notifications operation
	GenerateNotificationOperation TargetOperation = "GENERATE_NOTIFICATION"
)

// ResourceType represents the type of resource the constraint is applicable to
type ResourceType string

const (
	// ApplicationResourceType represents the application resource type
	ApplicationResourceType ResourceType = "APPLICATION"
	// RuntimeResourceType represents the runtime resource type
	RuntimeResourceType ResourceType = "RUNTIME"
	// RuntimeContextResourceType represents the runtime context resource type
	RuntimeContextResourceType ResourceType = "RUNTIME_CONTEXT"
	// TenantResourceType represents the tenant resource type
	TenantResourceType ResourceType = "TENANT"
	// FormationResourceType represents the formation resource type
	FormationResourceType ResourceType = "FORMATION"
)

// FormationConstraintScope defines the scope of the constraint
type FormationConstraintScope string

const (
	// GlobalFormationConstraintScope denotes the constraint is not bound to any formation type
	GlobalFormationConstraintScope FormationConstraintScope = "GLOBAL"
	// FormationTypeFormationConstraintScope denotes the constraint is applicable only to formations of the specified formation type
	FormationTypeFormationConstraintScope FormationConstraintScope = "FORMATION_TYPE"
)

// FormationConstraintInput represents the input for creating FormationConstraint
type FormationConstraintInput struct {
	Name            string
	ConstraintType  FormationConstraintType
	TargetOperation TargetOperation
	Operator        string
	ResourceType    ResourceType
	ResourceSubtype string
	InputTemplate   string
	ConstraintScope FormationConstraintScope
}

// FormationConstraint represents the constraint entity
type FormationConstraint struct {
	ID              string
	Name            string
	ConstraintType  FormationConstraintType
	TargetOperation TargetOperation
	Operator        string
	ResourceType    ResourceType
	ResourceSubtype string
	InputTemplate   string
	ConstraintScope FormationConstraintScope
}
