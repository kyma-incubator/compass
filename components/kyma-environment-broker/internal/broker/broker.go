package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

const (
	kymaServiceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"

	// time delay after which the instance becomes obsolete in the process of polling for last operation
	delayInstanceTime = 3 * time.Hour
)

//go:generate mockery -name=OptionalComponentNamesProvider -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=InputBuilderForPlan -output=automock -outpkg=automock -case=underscore

// OptionalComponentNamesProvider provides optional components names
type OptionalComponentNamesProvider interface {
	GetAllOptionalComponentsNames() []string
}

type DirectorClient interface {
	GetConsoleURL(accountID, runtimeID string) (string, error)
}

type StructDumper interface {
	Dump(value ...interface{})
}

// ProvisioningConfig holds all configurations connected with Provisioner API
type ProvisioningConfig struct {
	URL             string
	GCPSecretName   string
	AzureSecretName string
	AWSSecretName   string
}

var planIDsMapping = map[string]string{
	"azure": azurePlanID,
	"gcp":   gcpPlanID,
}

// Config represents configuration for broker
type Config struct {
	EnablePlans EnablePlans `envconfig:"default=azure"`
}

// EnablePlans defines the plans that should be available for provisioning
type EnablePlans []string

// Unmarshal provides custom parsing of Log Level.
// Implements envconfig.Unmarshal interface.
func (m *EnablePlans) Unmarshal(in string) error {
	plans := strings.Split(in, ",")
	for _, name := range plans {
		if _, exists := planIDsMapping[name]; !exists {
			return errors.Errorf("unrecognized %v plan name ", name)
		}
	}

	*m = plans
	return nil
}

// KymaEnvBroker implements the Kyma Environment Broker
type KymaEnvBroker struct {
	dumper             StructDumper
	provisioningCfg    ProvisioningConfig
	provisionerClient  provisioner.Client
	instancesStorage   storage.Instances
	builderFactory     InputBuilderForPlan
	DirectorClient     DirectorClient
	optionalComponents OptionalComponentNamesProvider
	enabledPlanIDs     map[string]struct{}
}

func New(cfg Config, pCli provisioner.Client, dCli DirectorClient, provisioningCfg ProvisioningConfig, instStorage storage.Instances, optComponentsSvc OptionalComponentNamesProvider,
	builderFactory InputBuilderForPlan, dumper StructDumper) (*KymaEnvBroker, error) {

	enabledPlanIDs := map[string]struct{}{}
	for _, planName := range cfg.EnablePlans {
		id := planIDsMapping[planName]
		enabledPlanIDs[id] = struct{}{}
	}

	return &KymaEnvBroker{
		provisionerClient:  pCli,
		DirectorClient:     dCli,
		dumper:             dumper,
		provisioningCfg:    provisioningCfg,
		instancesStorage:   instStorage,
		enabledPlanIDs:     enabledPlanIDs,
		builderFactory:     builderFactory,
		optionalComponents: optComponentsSvc,
	}, nil
}

// Services gets the catalog of services offered by the service broker
//   GET /v2/catalog
func (b *KymaEnvBroker) Services(ctx context.Context) ([]domain.Service, error) {
	var availableServicePlans []domain.ServicePlan

	for _, plan := range plans {
		// filter out not enabled plans
		if _, exists := b.enabledPlanIDs[plan.planDefinition.ID]; !exists {
			continue
		}
		p := plan.planDefinition
		err := json.Unmarshal(plan.provisioningRawSchema, &p.Schemas.Instance.Create.Parameters)
		b.addComponentsToSchema(&p.Schemas.Instance.Create.Parameters)
		if err != nil {
			b.dumper.Dump("Could not decode provisioning schema:", err.Error())
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
		SetProvisioningConfig(b.provisioningCfg)

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

// Update modifies an existing service instance
//  PATCH /v2/service_instances/{instance_id}
func (b *KymaEnvBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	b.dumper.Dump("Update instanceID:", instanceID)
	b.dumper.Dump("Update details:", details)
	b.dumper.Dump("Update asyncAllowed:", asyncAllowed)

	return domain.UpdateServiceSpec{}, nil
}

// LastOperation fetches last operation state for a service instance
//   GET /v2/service_instances/{instance_id}/last_operation
func (b *KymaEnvBroker) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	b.dumper.Dump("LastOperation instanceID:", instanceID)
	b.dumper.Dump("LastOperation details:", details)

	instance, err := b.instancesStorage.GetByID(instanceID)
	if err != nil {
		return domain.LastOperation{}, errors.Wrapf(err, "while getting instance from storage")
	}
	_, err = url.ParseRequestURI(instance.DashboardURL)
	if err == nil {
		return domain.LastOperation{
			State:       domain.Succeeded,
			Description: "Dashboard URL already exists in the instance",
		}, nil
	}

	status, err := b.provisionerClient.RuntimeOperationStatus(instance.GlobalAccountID, details.OperationData)
	if err != nil {
		b.dumper.Dump("Provisioner client returns error on runtime operation status call: ", err)
		return domain.LastOperation{}, errors.Wrapf(err, "while getting last operation")
	}
	b.dumper.Dump("Got status:", status)

	var lastOpStatus domain.LastOperationState
	var msg string
	if status.Message != nil {
		msg = *status.Message
	}

	switch status.State {
	case gqlschema.OperationStateSucceeded:
		operationStatus, directorMsg := b.handleDashboardURL(instance)
		if directorMsg != "" {
			msg = directorMsg
		}
		lastOpStatus = operationStatus
	case gqlschema.OperationStateInProgress:
		lastOpStatus = domain.InProgress
	case gqlschema.OperationStatePending:
		lastOpStatus = domain.InProgress
	case gqlschema.OperationStateFailed:
		lastOpStatus = domain.Failed
	}

	return domain.LastOperation{
		State:       lastOpStatus,
		Description: msg,
	}, nil
}

func (b *KymaEnvBroker) handleDashboardURL(instance *internal.Instance) (domain.LastOperationState, string) {
	b.dumper.Dump("Get dashboard url for instance ID: ", instance.InstanceID)

	dashboardURL, err := b.DirectorClient.GetConsoleURL(instance.GlobalAccountID, instance.RuntimeID)
	if director.IsTemporaryError(err) {
		b.dumper.Dump("DirectorClient cannot get Console URL (temporary): ", err.Error())
		state, msg := b.checkInstanceOutdated(instance)
		return state, fmt.Sprintf("cannot get URL from director: %s", msg)
	}
	if err != nil {
		b.dumper.Dump("DirectorClient cannot get Console URL: ", err.Error())
		return domain.Failed, fmt.Sprintf("cannot get URL from director: %s", err.Error())
	}

	instance.DashboardURL = dashboardURL
	err = b.instancesStorage.Update(*instance)
	if err != nil {
		b.dumper.Dump(fmt.Sprintf("Instance storage cannot update instance: %s", err))
		state, msg := b.checkInstanceOutdated(instance)
		return state, fmt.Sprintf("cannot update instance in storage: %s", msg)
	}

	return domain.Succeeded, ""
}

func (b *KymaEnvBroker) checkInstanceOutdated(instance *internal.Instance) (domain.LastOperationState, string) {
	addTime := instance.CreatedAt.Add(delayInstanceTime)
	subTime := time.Now().Sub(addTime)

	if subTime > 0 {
		// after delayInstanceTime Instance last operation is marked as failed
		b.dumper.Dump(fmt.Sprintf("Cannot get Dashboard URL for instance %s", instance.InstanceID))
		return domain.Failed, "instance is out of date"
	}

	return domain.InProgress, "action can be processed again"
}

// Bind creates a new service binding
//   PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	b.dumper.Dump("Bind instanceID:", instanceID)
	b.dumper.Dump("Bind details:", details)
	b.dumper.Dump("Bind asyncAllowed:", asyncAllowed)

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
	b.dumper.Dump("Unbind instanceID:", instanceID)
	b.dumper.Dump("Unbind details:", details)
	b.dumper.Dump("Unbind asyncAllowed:", asyncAllowed)

	unbind := domain.UnbindSpec{}
	return unbind, nil
}

// GetBinding fetches an existing service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	b.dumper.Dump("GetBinding instanceID:", instanceID)
	b.dumper.Dump("GetBinding bindingID:", bindingID)

	spec := domain.GetBindingSpec{}
	return spec, nil
}

// LastBindingOperation fetches last operation state for a service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation
func (b *KymaEnvBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	b.dumper.Dump("LastBindingOperation instanceID:", instanceID)
	b.dumper.Dump("LastBindingOperation bindingID:", bindingID)
	b.dumper.Dump("LastBindingOperation details:", details)

	op := domain.LastOperation{}
	return op, nil
}

func (b *KymaEnvBroker) addComponentsToSchema(schema *map[string]interface{}) {
	props := (*schema)["properties"].(map[string]interface{})
	props["components"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
			"enum": b.optionalComponents.GetAllOptionalComponentsNames(),
		},
	}
}
