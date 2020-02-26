package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

type ProvisionEndpoint struct {
	instancesStorage  storage.Instances
	builderFactory    InputBuilderForPlan
	provisioningCfg   ProvisioningConfig
	provisionerClient provisioner.Client
	dumper            StructDumper
	enabledPlanIDs    map[string]struct{}
}

// ProvisioningConfig holds all configurations connected with Provisioner API
type ProvisioningConfig struct {
	URL             string
	//GCPSecretName   string
	//AzureSecretName string
	//AWSSecretName   string
}

func NewProvision(cfg Config, instancesStorage storage.Instances, builderFactory InputBuilderForPlan, provisioningCfg ProvisioningConfig, provisionerClient provisioner.Client, dumper StructDumper) *ProvisionEndpoint {
	enabledPlanIDs := map[string]struct{}{}
	for _, planName := range cfg.EnablePlans {
		id := planIDsMapping[planName]
		enabledPlanIDs[id] = struct{}{}
	}

	return &ProvisionEndpoint{
		instancesStorage:  instancesStorage,
		builderFactory:    builderFactory,
		provisioningCfg:   provisioningCfg,
		provisionerClient: provisionerClient,
		dumper:            dumper,
		enabledPlanIDs:    enabledPlanIDs,
	}
}

// Provision creates a new service instance
//   PUT /v2/service_instances/{instance_id}
func (b *ProvisionEndpoint) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (domain.ProvisionedServiceSpec, error) {
	b.dumper.Dump("Provision instanceID:", instanceID)
	b.dumper.Dump("Provision details:", details)
	b.dumper.Dump("Provision asyncAllowed:", asyncAllowed)

	// unmarshall ERS context
	var ersContext internal.ERSContext
	err := json.Unmarshal(details.RawContext, &ersContext)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, errors.Wrap(err, "while decoding context")
	}
	if ersContext.GlobalAccountID == "" {
		return domain.ProvisionedServiceSpec{}, errors.New("GlobalAccountID parameter cannot be empty")
	}
	b.dumper.Dump("ERS context:", ersContext)

	if details.ServiceID != kymaServiceID {
		return domain.ProvisionedServiceSpec{}, errors.New("service_id not recognized")
	}

	// unmarshall provisioning parameters
	var parameters internal.ProvisioningParametersDTO
	err = json.Unmarshal(details.RawParameters, &parameters)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not read parameters, instanceID %s", instanceID))
	}

	if _, exists := b.enabledPlanIDs[details.PlanID]; !exists {
		return domain.ProvisionedServiceSpec{}, errors.Errorf("Plan ID %q is not recognized", details.PlanID)
	}

	// create input parameters according to selected plan
	inputBuilder, found := b.builderFactory.ForPlan(details.PlanID)
	if !found {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("The plan ID not known, instanceID %s, planID: %s", instanceID, details.PlanID))
	}
	inputBuilder.
		SetERSContext(ersContext).
		SetProvisioningParameters(parameters).
		SetProvisioningConfig(b.provisioningCfg).
		SetInstanceID(instanceID)

	input, err := inputBuilder.Build()
	if err != nil {
		return domain.ProvisionedServiceSpec{}, errors.Wrap(err, "while building provisioning inputToReturn")
	}

	b.dumper.Dump("Created provisioning input:", input)
	resp, err := b.provisionerClient.ProvisionRuntime(ersContext.GlobalAccountID, input)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not provision runtime, instanceID %s", instanceID))
	}
	if resp.RuntimeID == nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusInternalServerError, fmt.Sprintf("could not provision runtime, runtime ID not provided (instanceID %s)", instanceID))
	}
	err = b.instancesStorage.Insert(internal.Instance{
		InstanceID:             instanceID,
		GlobalAccountID:        ersContext.GlobalAccountID,
		RuntimeID:              *resp.RuntimeID,
		ServiceID:              details.ServiceID,
		ServicePlanID:          details.PlanID,
		DashboardURL:           "",
		ProvisioningParameters: string(details.RawParameters),
	})
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusInternalServerError, fmt.Sprintf("could not save instance, instanceID %s", instanceID))
	}

	spec := domain.ProvisionedServiceSpec{
		IsAsync:       true,
		OperationData: *resp.ID,
		DashboardURL:  "",
	}
	b.dumper.Dump("Returned provisioned service spec:", spec)

	return spec, nil
}
