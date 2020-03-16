package broker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

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
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(fmt.Errorf("instance not found"), http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	_, err = b.provisionerClient.DeprovisionRuntime(instance.GlobalAccountID, instance.RuntimeID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	return domain.DeprovisionServiceSpec{
		IsAsync: false,
	}, nil
}
