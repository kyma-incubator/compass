package deprovisioning

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type OperationManager struct {
	storage storage.Operations
}

func NewOperationManager(storage storage.Operations) *OperationManager {
	return &OperationManager{
		storage: storage,
	}
}

// OperationSucceeded marks the operation as succeeded and only repeats it if there is a storage error
func (om *OperationManager) OperationSucceeded(operation internal.DeprovisioningOperation, description string) (internal.DeprovisioningOperation, time.Duration, error) {
	updatedOperation, repeat := om.update(operation, domain.Succeeded, description)
	// repeat in case of storage error
	if repeat != 0 {
		return updatedOperation, repeat, nil
	}

	return updatedOperation, 0, nil
}

// OperationFailed marks the operation as failed and only repeats it if there is a storage error
func (om *OperationManager) OperationFailed(operation internal.DeprovisioningOperation, description string) (internal.DeprovisioningOperation, time.Duration, error) {
	updatedOperation, repeat := om.update(operation, domain.Failed, description)
	// repeat in case of storage error
	if repeat != 0 {
		return updatedOperation, repeat, nil
	}

	return updatedOperation, 0, errors.New(description)
}

// UpdateOperation updates a given operation
func (om *OperationManager) UpdateOperation(operation internal.DeprovisioningOperation) (internal.DeprovisioningOperation, time.Duration, error) {
	updatedOperation, err := om.storage.UpdateDeprovisioningOperation(operation)
	if err != nil {
		return operation, 1 * time.Minute, nil
	}
	return *updatedOperation, 0, nil
}

// InsertOperation stores operation in database
func (om *OperationManager) InsertOperation(operation internal.DeprovisioningOperation) (internal.DeprovisioningOperation, time.Duration, error) {
	err := om.storage.InsertDeprovisioningOperation(operation)
	if err != nil {
		return operation, 1 * time.Minute, nil
	}
	return operation, 0, nil
}

// RetryOperationOnce retries the operation once and fails the operation when call second time
func (om *OperationManager) RetryOperationOnce(operation internal.DeprovisioningOperation, errorMessage string, wait time.Duration, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	return om.RetryOperation(operation, errorMessage, wait, wait+1, log)
}

// RetryOperation retries an operation for at maxTime in retryInterval steps and fails the operation if retrying failed
func (om *OperationManager) RetryOperation(operation internal.DeprovisioningOperation, errorMessage string, retryInterval time.Duration, maxTime time.Duration, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	since := time.Since(operation.UpdatedAt)

	log.Infof("Retrying for %s in %s steps", maxTime.String(), retryInterval.String())
	if since < maxTime {
		return operation, retryInterval, nil
	}
	log.Errorf("Aborting after %s of failing retries", maxTime.String())
	return om.OperationFailed(operation, errorMessage)
}

func (om *OperationManager) update(operation internal.DeprovisioningOperation, state domain.LastOperationState, description string) (internal.DeprovisioningOperation, time.Duration) {
	operation.State = state
	operation.Description = fmt.Sprintf("%s : %s", operation.Description, description)

	updatedOperation, err := om.storage.UpdateDeprovisioningOperation(operation)
	// repeat if there is a problem with the storage
	if err != nil {
		return operation, 1 * time.Minute
	}

	return *updatedOperation, 0
}
