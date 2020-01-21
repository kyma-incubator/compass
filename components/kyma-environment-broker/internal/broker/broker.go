package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

const (
	kymaServiceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"

	fixedDummyURL = "https://dummy.dashboard.com"
)

// OptionalComponentNamesProvider provides optional components names
type OptionalComponentNamesProvider interface {
	GetOptionalComponentNames() []string
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
	Dumper *Dumper

	Config            ProvisioningConfig
	ProvisionerClient provisioner.Client

	InstancesStorage   storage.Instances
	optionalComponents OptionalComponentNamesProvider
}

var enabledPlanIDs = map[string]struct{}{
	azurePlanID: {},
	// add plan IDs which must be enabled
}

func NewBroker(pCli provisioner.Client, cfg ProvisioningConfig, instStorage storage.Instances) (*KymaEnvBroker, error) {
	dumper, err := NewDumper()
	if err != nil {
		return nil, err
	}

	return &KymaEnvBroker{
		ProvisionerClient:  pCli,
		Dumper:             dumper,
		Config:             cfg,
		InstancesStorage:   instStorage,
		optionalComponents: optionalComponentProvider{},
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
		b.addComponentsToSchema(&p.Schemas.Instance.Create.Parameters)
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
	var ersContext internal.ERSContext
	err := json.Unmarshal(details.RawContext, &ersContext)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, errors.Wrap(err, "while decoding context")
	}
	if ersContext.GlobalAccountID == "" {
		return domain.ProvisionedServiceSpec{}, errors.New("GlobalAccountID parameter cannot be empty")
	}
	b.Dumper.Dump("ERS context:", ersContext)

	if details.ServiceID != kymaServiceID {
		return domain.ProvisionedServiceSpec{}, errors.New("service_id not recognized")
	}

	// unmarshall provisioning parameters
	var parameters internal.ProvisioningParametersDTO
	err = json.Unmarshal(details.RawParameters, &parameters)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not read parameters, instanceID %s", instanceID))
	}
	b.Dumper.Dump("Provision parameters:", parameters)

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
	resp, err := b.ProvisionerClient.ProvisionRuntime(ersContext.GlobalAccountID, instanceID, *input)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusBadRequest, fmt.Sprintf("could not provision runtime, instanceID %s", instanceID))
	}
	if resp.RuntimeID == nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponseBuilder(err, http.StatusInternalServerError, fmt.Sprintf("could not provision runtime, runtime ID not provided (instanceID %s)", instanceID))
	}
	err = b.InstancesStorage.Insert(internal.Instance{
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
	b.Dumper.Dump("Returned provisioned service spec:", spec)

	return spec, nil
}

// Deprovision deletes an existing service instance
//  DELETE /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	b.Dumper.Dump("Deprovision instanceID:", instanceID)
	b.Dumper.Dump("Deprovision details:", details)
	b.Dumper.Dump("Deprovision asyncAllowed:", asyncAllowed)

	instance, err := b.InstancesStorage.GetByID(instanceID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.NewFailureResponseBuilder(fmt.Errorf("instance not found"), http.StatusBadRequest, fmt.Sprintf("could not deprovision runtime, instanceID %s", instanceID))
	}

	opID, err := b.ProvisionerClient.DeprovisionRuntime(instance.GlobalAccountID, instance.RuntimeID)
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

	inst, err := b.InstancesStorage.GetByID(instanceID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, errors.Wrapf(err, "while getting instance from storage")
	}

	decodedParams := make(map[string]interface{})
	err = json.Unmarshal([]byte(inst.ProvisioningParameters), &decodedParams)
	if err != nil {
		b.Dumper.Dump("unable to decode instance parameters for instanceID: ", instanceID)
		b.Dumper.Dump("  parameters: ", inst.ProvisioningParameters)
	}

	spec := domain.GetInstanceDetailsSpec{
		ServiceID:    inst.ServiceID,
		PlanID:       inst.ServicePlanID,
		DashboardURL: inst.DashboardURL,
		Parameters:   decodedParams,
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

	instance, err := b.InstancesStorage.GetByID(instanceID)
	if err != nil {
		return domain.LastOperation{}, errors.Wrapf(err, "while getting instance from storage")
	}

	status, err := b.ProvisionerClient.RuntimeOperationStatus(instance.GlobalAccountID, details.OperationData)
	if err != nil {
		b.Dumper.Dump("Got error: ", err)
		return domain.LastOperation{}, errors.Wrapf(err, "while getting last operation")
	}
	b.Dumper.Dump("Got status:", status)

	var lastOpStatus domain.LastOperationState
	switch status.State {
	case gqlschema.OperationStateSucceeded:
		lastOpStatus = domain.Succeeded

		// todo: this is a temporary solution until the dashboard url is saved in the director
		b.Dumper.Dump("Saving dummy dashboard url for instance ID: ", instanceID)
		inst, err := b.InstancesStorage.GetByID(instanceID)
		if err != nil {
			b.Dumper.Dump("Got error: ", err)
		}
		inst.DashboardURL = fixedDummyURL
		err = b.InstancesStorage.Update(*inst)
		if err != nil {
			b.Dumper.Dump("Got error: ", err)
		}

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
	b.Dumper.Dump("Bind instanceID:", instanceID)
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

func (b *KymaEnvBroker) addComponentsToSchema(schema *map[string]interface{}) {
	props := (*schema)["properties"].(map[string]interface{})
	props["components"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
			"enum": b.optionalComponents.GetOptionalComponentNames(),
		},
	}
}

// todo: will be replaced by the real implementation
type optionalComponentProvider struct {
}

func (optionalComponentProvider) GetOptionalComponentNames() []string {
	return []string{"monitoring", "kiali", "loki", "jaeger"}
}
