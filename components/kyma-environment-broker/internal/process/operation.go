package process

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type ProvisionOperationManager struct {
	storage storage.Provisioning
}

func NewProvisionOperationManager(storage storage.Operations) *ProvisionOperationManager {
	return &ProvisionOperationManager{storage: storage}
}

// OperationSucceeded marks the operation as succeeded and only repeats it if there is a storage error
func (om *ProvisionOperationManager) OperationSucceeded(operation internal.ProvisioningOperation, description string) (internal.ProvisioningOperation, time.Duration, error) {
	updatedOperation, repeat := om.update(operation, domain.Succeeded, description)
	// repeat in case of storage error
	if repeat != 0 {
		return updatedOperation, repeat, nil
	}

	return updatedOperation, 0, nil
}

// OperationFailed marks the operation as failed and only repeats it if there is a storage error
func (om *ProvisionOperationManager) OperationFailed(operation internal.ProvisioningOperation, description string) (internal.ProvisioningOperation, time.Duration, error) {
	updatedOperation, repeat := om.update(operation, domain.Failed, description)
	// repeat in case of storage error
	if repeat != 0 {
		return updatedOperation, repeat, nil
	}

	return updatedOperation, 0, errors.New(description)
}

// UpdateOperation updates a given operation
func (om *ProvisionOperationManager) UpdateOperation(operation internal.ProvisioningOperation) (internal.ProvisioningOperation, time.Duration) {
	updatedOperation, err := om.storage.UpdateProvisioningOperation(operation)
	if err != nil {
		return operation, 1 * time.Minute
	}
	return *updatedOperation, 0
}

// RetryOperationOnce retries the operation once and fails the operation when call second time
func (om *ProvisionOperationManager) RetryOperationOnce(operation internal.ProvisioningOperation, errorMessage string, wait time.Duration, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	return om.RetryOperation(operation, errorMessage, wait, wait+1, log)
}

// RetryOperation retries an operation for at maxTime in retryInterval steps and fails the operation if retrying failed
func (om *ProvisionOperationManager) RetryOperation(operation internal.ProvisioningOperation, errorMessage string, retryInterval time.Duration, maxTime time.Duration, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	since := time.Since(operation.UpdatedAt)

	log.Infof("Retrying for %s in %s steps", maxTime.String(), retryInterval.String())
	if since < maxTime {
		return operation, retryInterval, nil
	}
	log.Errorf("Aborting after %s of failing retries", maxTime.String())
	return om.OperationFailed(operation, errorMessage)
}

func (om *ProvisionOperationManager) update(operation internal.ProvisioningOperation, state domain.LastOperationState, description string) (internal.ProvisioningOperation, time.Duration) {
	operation.State = state
	operation.Description = fmt.Sprintf("%s : %s", operation.Description, description)

	return om.UpdateOperation(operation)
}
