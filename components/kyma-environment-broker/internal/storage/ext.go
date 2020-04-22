package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/predicate"
)

type Instances interface {
	FindAllJoinedWithOperations(prct ...predicate.Predicate) ([]internal.InstanceWithOperation, error)
	GetByID(instanceID string) (*internal.Instance, error)
	Insert(instance internal.Instance) error
	Update(instance internal.Instance) error
	Delete(instanceID string) error
}

type Operations interface {
	Provisioning
	Deprovisioning

	GetOperationByID(operationID string) (*internal.Operation, error)
	GetOperationsInProgressByType(operationType dbmodel.OperationType) ([]internal.Operation, error)
}

type Provisioning interface {
	InsertProvisioningOperation(operation internal.ProvisioningOperation) error
	GetProvisioningOperationByID(operationID string) (*internal.ProvisioningOperation, error)
	GetProvisioningOperationByInstanceID(instanceID string) (*internal.ProvisioningOperation, error)
	UpdateProvisioningOperation(operation internal.ProvisioningOperation) (*internal.ProvisioningOperation, error)
}

type Deprovisioning interface {
	InsertDeprovisioningOperation(operation internal.DeprovisioningOperation) error
	GetDeprovisioningOperationByID(operationID string) (*internal.DeprovisioningOperation, error)
	GetDeprovisioningOperationByInstanceID(instanceID string) (*internal.DeprovisioningOperation, error)
	UpdateDeprovisioningOperation(operation internal.DeprovisioningOperation) (*internal.DeprovisioningOperation, error)
}

type LMSTenants interface {
	FindTenantByName(name, region string) (internal.LMSTenant, bool, error)
	InsertTenant(tenant internal.LMSTenant) error
}
