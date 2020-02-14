package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
)

type Instances interface {
	GetByID(instanceID string) (*internal.Instance, error)
	Insert(instance internal.Instance) error
	Update(instance internal.Instance) error
	Delete(instanceID string) error
}

type Operations interface {
	InsertProvisioningOperation(operation internal.ProvisioningOperation) error
	GetProvisioningOperationByID(operationID string) (*internal.ProvisioningOperation, error)
	GetProvisioningOperationByInstanceID(instanceID string) (*internal.ProvisioningOperation, error)
	UpdateProvisioningOperation(operation internal.ProvisioningOperation) (*internal.ProvisioningOperation, error)
	GetOperation(operationID string) (*internal.Operation, error)
	GetOperationsInProgressByType(operationType dbsession.OperationType) ([]internal.Operation, error)
}

type LMSTenants interface {
	FindTenantByName(name, region string) (internal.LMSTenant, bool, error)
	InsertTenant(tenant internal.LMSTenant) error
}
