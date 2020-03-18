package broker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/sirupsen/logrus"
)

type DeprovisionEndpoint struct {
	log logrus.FieldLogger

	instancesStorage  storage.Instances
	provisionerClient provisioner.Client
}

func NewDeprovision(instancesStorage storage.Instances, provisionerClient provisioner.Client, log logrus.FieldLogger) *DeprovisionEndpoint {
	return &DeprovisionEndpoint{
		log:               log,
		instancesStorage:  instancesStorage,
		provisionerClient: provisionerClient,
	}
}

// Deprovision deletes an existing service instance
//  DELETE /v2/service_instances/{instance_id}
func (b *DeprovisionEndpoint) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	logger := b.log.WithField("instanceID", instanceID)
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

	logger.Infof("deprovision runtime: runtimeID=%s, globalAccountID=%s", instance.RuntimeID, instance.GlobalAccountID)
	_, err = b.provisionerClient.DeprovisionRuntime(instance.GlobalAccountID, instance.RuntimeID)
	if err != nil {
		logger.Errorf("unable to deprovision runtime: %s", err)
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponse(err, http.StatusInternalServerError, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	err = b.instancesStorage.Delete(instanceID)
	if err != nil {
		logger.Errorf("unable to delete instance: %s", err)
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponse(err, http.StatusInternalServerError, fmt.Sprintf("could not delete instance, instanceID %s", instanceID))
	}

	return domain.DeprovisionServiceSpec{
		IsAsync: false,
	}, nil
}
