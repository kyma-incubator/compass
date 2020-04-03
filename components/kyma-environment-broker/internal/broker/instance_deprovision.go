package broker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/sirupsen/logrus"
)

type DeprovisionEndpoint struct {
	log logrus.FieldLogger

	instancesStorage  storage.Instances
	operationsStorage storage.Deprovisioning

	queue Queue
}

func NewDeprovision(instancesStorage storage.Instances, operationsStorage storage.Operations, q Queue, log logrus.FieldLogger) *DeprovisionEndpoint {
	return &DeprovisionEndpoint{
		log:               log,
		instancesStorage:  instancesStorage,
		operationsStorage: operationsStorage,

		queue: q,
	}
}

// Deprovision deletes an existing service instance
//  DELETE /v2/service_instances/{instance_id}
func (b *DeprovisionEndpoint) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	operationID := uuid.New().String()
	logger := b.log.WithField("instanceID", instanceID).WithField("operationID", operationID)
	logger.Infof("Deprovisioning triggered, details: %+v", details)

	instance, err := b.instancesStorage.GetByID(instanceID)
	switch {
	case err == nil:
	case dberr.IsNotFound(err):
		logger.Warn("instance does not exists")
		return domain.DeprovisionServiceSpec{}, apiresponses.ErrInstanceDoesNotExist // HTTP 410 GONE
	default:
		logger.Errorf("unable to get instance from a storage: %s", err)
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponse(fmt.Errorf("unable to get instance from the storage"), http.StatusInternalServerError, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	// check if operation with the same instance ID is already created
	existingOperation, errStorage := b.operationsStorage.GetDeprovisioningOperationByInstanceID(instanceID)
	switch {
	case errStorage != nil && !dberr.IsNotFound(errStorage):
		logger.Errorf("cannot get existing operation from storage %s", errStorage)
		return domain.DeprovisionServiceSpec{}, errors.New("cannot get existing operation from storage")
	case existingOperation != nil && !dberr.IsNotFound(errStorage):
		if existingOperation.State == domain.Failed {
			// reprocess operation again
			existingOperation.State = domain.InProgress
			_, err = b.operationsStorage.UpdateDeprovisioningOperation(*existingOperation)
			if err != nil {
				return domain.DeprovisionServiceSpec{}, errors.New("cannot update existing operation")
			}
			logger.Infof("Reprocessing failed deprovisioning of runtime: runtimeID=%s, globalAccountID=%s", instance.RuntimeID, instance.GlobalAccountID)
			b.queue.Add(operationID)
		}
		// return existing operation
		return domain.DeprovisionServiceSpec{
			IsAsync:       true,
			OperationData: existingOperation.ID,
		}, nil
	}
	// create and save new operation
	operation, err := internal.NewDeprovisioningOperationWithID(operationID, instanceID)
	if err != nil {
		logger.Errorf("cannot create new operation: %s", err)
		return domain.DeprovisionServiceSpec{}, errors.New("cannot create new operation")
	}
	err = b.operationsStorage.InsertDeprovisioningOperation(operation)
	if err != nil {
		logger.Errorf("cannot save operation: %s", err)
		return domain.DeprovisionServiceSpec{}, errors.New("cannot save operation")
	}

	logger.Infof("Deprovisioning runtime: runtimeID=%s, globalAccountID=%s", instance.RuntimeID, instance.GlobalAccountID)

	b.queue.Add(operationID)

	return domain.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: operationID,
	}, nil
}
