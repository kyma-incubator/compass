package broker

import (
	"context"
	"encoding/json"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

// GetInstance fetches information about a service instance
//   GET /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) GetInstance(ctx context.Context, instanceID string) (domain.GetInstanceDetailsSpec, error) {
	b.dumper.Dump("GetInstance instanceID:", instanceID)

	inst, err := b.instancesStorage.GetByID(instanceID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, errors.Wrapf(err, "while getting instance from storage")
	}

	decodedParams := make(map[string]interface{})
	err = json.Unmarshal([]byte(inst.ProvisioningParameters), &decodedParams)
	if err != nil {
		b.dumper.Dump("unable to decode instance parameters for instanceID: ", instanceID)
		b.dumper.Dump("  parameters: ", inst.ProvisioningParameters)
	}

	spec := domain.GetInstanceDetailsSpec{
		ServiceID:    inst.ServiceID,
		PlanID:       inst.ServicePlanID,
		DashboardURL: inst.DashboardURL,
		Parameters:   decodedParams,
	}
	return spec, nil
}
