package dbmodel

import "time"

// OperationType defines the possible types of an asynchronous operation to a broker.
type OperationType string

const (
	// OperationTypeProvision means provisioning OperationType
	OperationTypeProvision OperationType = "provision"
	// OperationTypeDeprovision means deprovision OperationType
	OperationTypeDeprovision OperationType = "deprovision"
	// OperationTypeUndefined means undefined OperationType
	OperationTypeUndefined OperationType = ""
)

type OperationDTO struct {
	ID        string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time

	InstanceID        string
	TargetOperationID string
	State             string
	Description       string

	Data string

	Type OperationType
}
