package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/domain"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
)

const (
	kymaServiceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"

	fixedDummyURL = "https://dummy.dashboard.com"
)

type ERSContext struct {
	TenantID        string `json:"tenant_id"`
	SubaccountID    string `json:"subaccount_id"`
	GlobalaccountID string `json:"globalaccount_id"`
}

// ProvisioningConfig holds all configurations connected with Provisioner API
type ProvisioningConfig struct {
	URL                 string
	SecretName          string
	GCPSecretName       string
	AzureSecretName     string
	AWSSecretName       string
	GardenerProjectName string
}

// KymaEnvBroker implements the Kyma Environment Broker
type KymaEnvBroker struct {
	Dumper            *Dumper
	ProvisionerClient provisioner.Client

	Config ProvisioningConfig

	// todo: remove after the storage is done
	intanceToRuntimeIDs map[string]string
}

type ProvisioningParameters struct {
	Name           string  `json:"name"`
	NodeCount      *int    `json:"nodeCount"`
	VolumeSizeGb   *int    `json:"volumeSizeGb"`
	MachineType    *string `json:"machineType"`
	Region         *string `json:"region"`
	Zone           *string `json:"zone"`
	AutoScalerMin  *int    `json:"autoScalerMin"`
	AutoScalerMax  *int    `json:"autoScalerMax"`
	MaxSurge       *int    `json:"maxSurge"`
	MaxUnavailable *int    `json:"maxUnavailable"`
}

var enabledPlanIDs = map[string]struct{}{
	azurePlanID: {},
	// add plan IDs which must be enabled
}

func NewBroker(pCli provisioner.Client, cfg ProvisioningConfig) (*KymaEnvBroker, error) {
	dumper, err := NewDumper()
	if err != nil {
		return nil, err
	}

	return &KymaEnvBroker{
		ProvisionerClient:   pCli,
		Dumper:              dumper,
		Config:              cfg,
		intanceToRuntimeIDs: map[string]string{},
	}, nil
}

// Services gets the catalog of services offered by the service broker
//   GET /v2/catalog
func (b *KymaEnvBroker) Services(ctx context.Context) ([]domain.Service, error) {
	availableServicePlans := []domain.ServicePlan{}

	for _, plan := range plans {
		// filter out not enabled plans
		if _, exists := enabledPlanIDs[plan.planDefinition.ID]; !exists {
			continue
		}
		p := plan.planDefinition
		err := json.Unmarshal(plan.provisioningRawSchema, &p.Schemas.Instance.Create.Parameters)
		if err != nil {
			b.Dumper.Dump("Could not decode provisioning schema:", err.Error())
			return nil, err
		}
		availableServicePlans = append(availableServicePlans, p)
	}

	return []domain.Service{
		{
			ID:          kymaServiceID,
			Name:        "kymaruntime",
			Description: "[EXPERIMENTAL] Service Class for Kyma Runtime",
			Bindable:    true,
			Plans:       availableServicePlans,
			Metadata: &domain.ServiceMetadata{
				DisplayName:         "Kyma Runtime",
				LongDescription:     "Kyma Runtime experimental service class",
				DocumentationUrl:    "kyma-project.io",
				ProviderDisplayName: "SAP",
			},
			Tags: []string{
				"SAP",
				"Kyma",
			},
		},
	}, nil
}

// Provision creates a new service instance
//   PUT /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (domain.ProvisionedServiceSpec, error) {
	b.Dumper.Dump("Provision instanceID:", instanceID)
	b.Dumper.Dump("Provision details:", details)
	b.Dumper.Dump("Provision asyncAllowed:", asyncAllowed)

	// unmarshall ERS context
	var ersContext ERSContext
	err := json.Unmarshal(details.RawContext, &ersContext)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, errors.Wrap(err, "while decoding context")
	}
	b.Dumper.Dump("ERS context:", ersContext)

	if details.ServiceID != kymaServiceID {
		return domain.ProvisionedServiceSpec{}, errors.New("service_id not recognized")
	}

	// unmarshall provisioning parameters
	var parameters ProvisioningParameters
	err = json.Unmarshal(details.RawParameters, &parameters)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not read parameters, instanceID %s", instanceID))
	}

	// create input parameters according to selected provider
	inputBuilder, found := NewInputBuilderForPlan(details.PlanID)
	if !found {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("The plan ID not known, instanceID %s, planID: %s", instanceID, details.PlanID))
	}
	inputBuilder.ApplyParameters(&parameters)
	// todo: ApplyERSContext()
	input := inputBuilder.ClusterConfigInput()

	// add values, which will be deprecated and replaced by other secret data provided by the caller
	switch details.PlanID {
	case azurePlanID:
		input.ClusterConfig.GardenerConfig.TargetSecret = b.Config.AzureSecretName
	default:
		return domain.ProvisionedServiceSpec{}, errors.Wrapf(err, "unknown Plan ID %s", details.PlanID)
	}
	input.ClusterConfig.GardenerConfig.ProjectName = b.Config.GardenerProjectName
	input.Credentials = &gqlschema.CredentialsInput{
		SecretName: b.Config.SecretName,
	}

	b.Dumper.Dump("Created provisioning input:", input)
	resp, err := b.ProvisionerClient.ProvisionRuntime(instanceID, *input)

	// todo: store in the storage
	if resp.RuntimeID == nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusInternalServerError, fmt.Sprintf("could not provision runtime, runtime ID not provided (instanceID %s)", instanceID))
	}
	b.registerRuntime(instanceID, *resp.RuntimeID)

	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not provision runtime, instanceID %s", instanceID))
	}

	spec := domain.ProvisionedServiceSpec{
		IsAsync:       true,
		OperationData: *resp.ID,
		DashboardURL:  fixedDummyURL,
	}
	b.Dumper.Dump("Returned provisioned service spec:", spec)

	return spec, nil
}

// Deprovision deletes an existing service instance
//  DELETE /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	b.Dumper.Dump("Deprovision instanceID:", instanceID)
	b.Dumper.Dump("Deprovision details:", details)
	b.Dumper.Dump("Deprovision asyncAllowed:", asyncAllowed)

	// todo: read from storage
	runtimeID, found := b.runtimeIDByInstance(instanceID)
	if !found {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(fmt.Errorf("instance not found"), http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	opID, err := b.ProvisionerClient.DeprovisionRuntime(runtimeID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	return domain.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: opID,
	}, nil
}

// GetInstance fetches information about a service instance
//   GET /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) GetInstance(ctx context.Context, instanceID string) (domain.GetInstanceDetailsSpec, error) {
	b.Dumper.Dump("GetInstance instanceID:", instanceID)

	runtimeStatus, err := b.ProvisionerClient.GCPRuntimeStatus(instanceID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, errors.Wrapf(err, "while calling for runtime status")
	}

	b.Dumper.Dump("GCP Runtime Status cluster config:", runtimeStatus.RuntimeConfiguration.ClusterConfig)
	b.Dumper.Dump("GCP Runtime Status Kyma config:", runtimeStatus.RuntimeConfiguration.KymaConfig)

	spec := domain.GetInstanceDetailsSpec{
		ServiceID:    kymaServiceID,
		PlanID:       azurePlanID, // todo: read the ID from the storage
		DashboardURL: fixedDummyURL,
	}
	return spec, nil
}

// Update modifies an existing service instance
//  PATCH /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	b.Dumper.Dump("Update instanceID:", instanceID)
	b.Dumper.Dump("Update details:", details)
	b.Dumper.Dump("Update asyncAllowed:", asyncAllowed)

	return domain.UpdateServiceSpec{}, nil
}

// LastOperation fetches last operation state for a service instance
//   GET /v2/service_instances/{instance_id}/last_operation
func (b *KymaEnvBroker) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	b.Dumper.Dump("LastOperation instanceID:", instanceID)
	b.Dumper.Dump("LastOperation details:", details)

	status, err := b.ProvisionerClient.RuntimeOperationStatus(details.OperationData)
	if err != nil {
		b.Dumper.Dump("Got error: ", err)
		return domain.LastOperation{}, errors.Wrapf(err, "while getting last operation")
	}
	b.Dumper.Dump("Got status:", status)

	var lastOpStatus domain.LastOperationState
	switch status.State {
	case gqlschema.OperationStateSucceeded:
		lastOpStatus = domain.Succeeded
	case gqlschema.OperationStateInProgress:
		lastOpStatus = domain.InProgress
	case gqlschema.OperationStatePending:
		lastOpStatus = domain.InProgress
	case gqlschema.OperationStateFailed:
		lastOpStatus = domain.Failed
	}
	msg := ""
	if status.Message != nil {
		msg = *status.Message
	}

	return domain.LastOperation{
		State:       lastOpStatus,
		Description: msg,
	}, nil
}

// Bind creates a new service binding
//   PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	litter.Dump("Bind instanceID:", instanceID)
	b.Dumper.Dump("Bind details:", details)
	b.Dumper.Dump("Bind asyncAllowed:", asyncAllowed)

	binding := domain.Binding{
		Credentials: map[string]interface{}{
			"host":     "test",
			"port":     "1234",
			"password": "nimda123",
		},
	}
	return binding, nil
}

// Unbind deletes an existing service binding
//   DELETE /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	b.Dumper.Dump("Unbind instanceID:", instanceID)
	b.Dumper.Dump("Unbind details:", details)
	b.Dumper.Dump("Unbind asyncAllowed:", asyncAllowed)

	unbind := domain.UnbindSpec{}
	return unbind, nil
}

// GetBinding fetches an existing service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	b.Dumper.Dump("GetBinding instanceID:", instanceID)
	b.Dumper.Dump("GetBinding bindingID:", bindingID)

	spec := domain.GetBindingSpec{}
	return spec, nil
}

// LastBindingOperation fetches last operation state for a service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation
func (b *KymaEnvBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	b.Dumper.Dump("LastBindingOperation instanceID:", instanceID)
	b.Dumper.Dump("LastBindingOperation bindingID:", bindingID)
	b.Dumper.Dump("LastBindingOperation details:", details)

	op := domain.LastOperation{}
	return op, nil
}

// todo: remove after the storage is done
func (b *KymaEnvBroker) registerRuntime(instnanceID string, runtimeID string) {
	b.intanceToRuntimeIDs[instnanceID] = runtimeID
}

// todo: remove after the storage is done
func (b *KymaEnvBroker) runtimeIDByInstance(instanceID string) (string, bool) {
	rID, found := b.intanceToRuntimeIDs[instanceID]
	return rID, found
}
