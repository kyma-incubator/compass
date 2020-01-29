package broker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

// Deprovision deletes an existing service instance
//  DELETE /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	b.dumper.Dump("Deprovision instanceID:", instanceID)
	b.dumper.Dump("Deprovision details:", details)
	b.dumper.Dump("Deprovision asyncAllowed:", asyncAllowed)

	instance, err := b.instancesStorage.GetByID(instanceID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(fmt.Errorf("instance not found"), http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	opID, err := b.provisionerClient.DeprovisionRuntime(instance.GlobalAccountID, instance.RuntimeID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	return domain.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: opID,
	}, nil
}
