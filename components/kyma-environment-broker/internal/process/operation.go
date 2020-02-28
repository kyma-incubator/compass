package process

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type OperationManager struct {
	storage storage.Operations
}

func NewOperationManager(storage storage.Operations) *OperationManager {
	return &OperationManager{storage: storage}
}

func (om *OperationManager) OperationSucceeded(operation internal.ProvisioningOperation, description string) (internal.ProvisioningOperation, time.Duration, error) {
	updatedOperation, repeat := om.update(operation, domain.Succeeded, description)
	if repeat != 0 {
		return updatedOperation, repeat, nil
	}

	return updatedOperation, 0, nil
}

func (om *OperationManager) OperationFailed(operation internal.ProvisioningOperation, description string) (internal.ProvisioningOperation, time.Duration, error) {
	updatedOperation, repeat := om.update(operation, domain.Failed, description)
	if repeat != 0 {
		return updatedOperation, repeat, nil
	}

	return updatedOperation, 0, errors.New(description)
}

func (om *OperationManager) UpdateOperation(operation internal.ProvisioningOperation) (internal.ProvisioningOperation, time.Duration) {
	updatedOperation, err := om.storage.UpdateProvisioningOperation(operation)
	if err != nil {
		return operation, 1 * time.Minute
	}

	return *updatedOperation, 0
}

func (om *OperationManager) update(operation internal.ProvisioningOperation, state domain.LastOperationState, description string) (internal.ProvisioningOperation, time.Duration) {
	operation.State = state
	operation.Description = fmt.Sprintf("%s : %s", operation.Description, description)

	updatedOperation, err := om.storage.UpdateProvisioningOperation(operation)
	if err != nil {
		return operation, 1 * time.Minute
	}

	return *updatedOperation, 0
}
