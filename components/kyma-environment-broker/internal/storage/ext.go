package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

type Instances interface {
	GetByID(instanceID string) (*internal.Instance, error)
	Insert(instance internal.Instance) error
	Update(instance internal.Instance) error
}

type Operations interface {
	InsertProvisioningOperation(operation internal.ProvisioningOperation) error
	GetProvisioningOperationByID(operationID string) (*internal.ProvisioningOperation, error)
	UpdateProvisioningOperation(operation internal.ProvisioningOperation) (*internal.ProvisioningOperation, error)
	GetOperation(operationID string) (*internal.Operation, error)
}
