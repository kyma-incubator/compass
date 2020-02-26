package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Queue -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=PlanValidator -output=automock -outpkg=automock -case=underscore

type (
	Queue interface {
		Add(operationId string)
	}

	PlanValidator interface {
		IsPlanSupport(planID string) bool
	}
)

type ProvisionEndpoint struct {
	operationsStorage storage.Operations
	queue             Queue
	builderFactory    PlanValidator
	dumper            StructDumper
	enabledPlanIDs    map[string]struct{}
}

func NewProvision(cfg Config, operationsStorage storage.Operations, q Queue, builderFactory PlanValidator, dumper StructDumper) *ProvisionEndpoint {
	enabledPlanIDs := map[string]struct{}{}
	for _, planName := range cfg.EnablePlans {
		id := planIDsMapping[planName]
		enabledPlanIDs[id] = struct{}{}
	}

	return &ProvisionEndpoint{
		operationsStorage: operationsStorage,
		queue:             q,
		builderFactory:    builderFactory,
		dumper:            dumper,
		enabledPlanIDs:    enabledPlanIDs,
	}
}

// Provision creates a new service instance
//   PUT /v2/service_instances/{instance_id}
func (b *ProvisionEndpoint) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (domain.ProvisionedServiceSpec, error) {
	// validation of incoming input
	ersContext, parameters, err := b.validate(details)
	if err != nil {
		errMsg := fmt.Sprintf("[instanceID: %s] %s", instanceID, err)
		return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponse(err, http.StatusBadRequest, errMsg)
	}

	provisioningParameters := internal.ProvisioningParameters{
		PlanID:     details.PlanID,
		ServiceID:  details.ServiceID,
		ErsContext: ersContext,
		Parameters: parameters,
	}

	// check if operation with instance ID already created
	existingOperation, errStorage := b.operationsStorage.GetProvisioningOperationByInstanceID(instanceID)
	switch {
	case errStorage != nil && !dberr.IsNotFound(errStorage):
		b.dumper.Dump("cannot get existing operation from storage", errStorage)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot get existing operation from storage")
	case existingOperation != nil && !dberr.IsNotFound(errStorage):
		return b.handleExistingOperation(existingOperation, provisioningParameters)
	}

	// create and save new operation
	operation, err := internal.NewProvisioningOperation(instanceID, provisioningParameters)
	if err != nil {
		b.dumper.Dump("cannot create new operation: ", err)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot create new operation")
	}

	err = b.operationsStorage.InsertProvisioningOperation(operation)
	if err != nil {
		b.dumper.Dump("cannot save operations: ", err)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot save operations")
	}

	// add new operation to queue
	b.queue.Add(operation.ID)

	return domain.ProvisionedServiceSpec{
		IsAsync:       true,
		OperationData: operation.ID,
	}, nil
}

func (b *ProvisionEndpoint) validate(details domain.ProvisionDetails) (internal.ERSContext, internal.ProvisioningParametersDTO, error) {
	var ersContext internal.ERSContext
	var parameters internal.ProvisioningParametersDTO

	if details.ServiceID != kymaServiceID {
		return ersContext, parameters, errors.New("service_id not recognized")
	}
	if _, exists := b.enabledPlanIDs[details.PlanID]; !exists {
		return ersContext, parameters, errors.Errorf("plan ID %q is not recognized", details.PlanID)
	}

	ersContext, err := b.extractERSContext(details)
	if err != nil {
		return ersContext, parameters, errors.Wrap(err, "while extracting ers context")
	}

	parameters, err = b.extractInputParameters(details)
	if err != nil {
		return ersContext, parameters, errors.Wrap(err, "while extracting input parameters")
	}

	found := b.builderFactory.IsPlanSupport(details.PlanID)
	if !found {
		return ersContext, parameters, errors.Errorf("the plan ID not known, planID: %s", details.PlanID)
	}

	return ersContext, parameters, nil
}

func (b *ProvisionEndpoint) extractERSContext(details domain.ProvisionDetails) (internal.ERSContext, error) {
	var ersContext internal.ERSContext
	err := json.Unmarshal(details.RawContext, &ersContext)
	if err != nil {
		return ersContext, errors.Wrap(err, "while decoding context")

	}
	if ersContext.GlobalAccountID == "" {
		return ersContext, errors.New("global accountID parameter cannot be empty")
	}

	return ersContext, nil
}

func (b *ProvisionEndpoint) extractInputParameters(details domain.ProvisionDetails) (internal.ProvisioningParametersDTO, error) {
	var parameters internal.ProvisioningParametersDTO
	err := json.Unmarshal(details.RawParameters, &parameters)
	if err != nil {
		return parameters, errors.Wrap(err, "while unmarshaling raw parameters")
	}

	return parameters, nil
}

func (b *ProvisionEndpoint) handleExistingOperation(operation *internal.ProvisioningOperation, input internal.ProvisioningParameters) (domain.ProvisionedServiceSpec, error) {
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		b.dumper.Dump("cannot get provisioning parameters from exist operation", err)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot get provisioning parameters from exist operation")
	}
	if pp.IsEqual(input) {
		return domain.ProvisionedServiceSpec{
			IsAsync:       true,
			AlreadyExists: true,
			OperationData: operation.ID,
		}, nil
	}

	err = errors.New("provisioning operation already exist")
	log := fmt.Sprintf("provisioning operation with InstanceID %s already exist", operation.InstanceID)
	return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponse(err, http.StatusConflict, log)
}
