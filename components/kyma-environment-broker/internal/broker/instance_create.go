package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"

	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	operationsStorage    storage.Provisioning
	instanceStorage      storage.Instances
	queue                Queue
	builderFactory       PlanValidator
	enabledPlanIDs       map[string]struct{}
	plansSchemaValidator PlansSchemaValidator
	kymaVerOnDemand      bool

	log logrus.FieldLogger
}

func NewProvision(cfg Config, operationsStorage storage.Operations, instanceStorage storage.Instances, q Queue, builderFactory PlanValidator, validator PlansSchemaValidator, kvod bool, log logrus.FieldLogger) *ProvisionEndpoint {
	enabledPlanIDs := map[string]struct{}{}
	for _, planName := range cfg.EnablePlans {
		id := planIDsMapping[planName]
		enabledPlanIDs[id] = struct{}{}
	}

	return &ProvisionEndpoint{
		plansSchemaValidator: validator,
		operationsStorage:    operationsStorage,
		instanceStorage:      instanceStorage,
		queue:                q,
		builderFactory:       builderFactory,
		log:                  log.WithField("service", "ProvisionEndpoint"),
		enabledPlanIDs:       enabledPlanIDs,
		kymaVerOnDemand:      kvod,
	}
}

// Provision creates a new service instance
//   PUT /v2/service_instances/{instance_id}
func (b *ProvisionEndpoint) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (domain.ProvisionedServiceSpec, error) {
	operationID := uuid.New().String()
	logger := b.log.WithField("instanceID", instanceID).WithField("operationID", operationID)
	logger.Infof("Provision called: planID=%s", details.PlanID)
	// validation of incoming input
	ersContext, parameters, err := b.validateAndExtract(details, logger)
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
	logger.Infof("Starting provisioning runtime: Name=%s, GlobalAccountID=%s, SubAccountID=%s", parameters.Name, ersContext.GlobalAccountID, ersContext.SubAccountID)
	logger.Infof("Runtime parameters: %+v", parameters)

	// check if operation with instance ID already created
	existingOperation, errStorage := b.operationsStorage.GetProvisioningOperationByInstanceID(instanceID)
	switch {
	case errStorage != nil && !dberr.IsNotFound(errStorage):
		logger.Errorf("cannot get existing operation from storage %s", errStorage)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot get existing operation from storage")
	case existingOperation != nil && !dberr.IsNotFound(errStorage):
		return b.handleExistingOperation(existingOperation, provisioningParameters, logger)
	}

	// create and save new operation
	operation, err := internal.NewProvisioningOperationWithID(operationID, instanceID, provisioningParameters)
	if err != nil {
		logger.Errorf("cannot create new operation: %s", err)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot create new operation")
	}

	err = b.operationsStorage.InsertProvisioningOperation(operation)
	if err != nil {
		logger.Errorf("cannot save operation: %s", err)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot save operation")
	}
	err = b.instanceStorage.Insert(internal.Instance{
		InstanceID:             instanceID,
		GlobalAccountID:        ersContext.GlobalAccountID,
		SubAccountID:           ersContext.SubAccountID,
		ServiceID:              provisioningParameters.ServiceID,
		ServiceName:            KymaServiceName,
		ServicePlanID:          provisioningParameters.PlanID,
		ServicePlanName:        Plans[provisioningParameters.PlanID].PlanDefinition.Name,
		ProvisioningParameters: operation.ProvisioningParameters,
	})
	if err != nil {
		logger.Errorf("cannot save instance in storage: %s", err)
		return domain.ProvisionedServiceSpec{}, errors.New("cannot save instance")
	}

	// add new operation to queue
	b.queue.Add(operation.ID)

	return domain.ProvisionedServiceSpec{
		IsAsync:       true,
		OperationData: operation.ID,
	}, nil
}

func (b *ProvisionEndpoint) validateAndExtract(details domain.ProvisionDetails, logger logrus.FieldLogger) (internal.ERSContext, internal.ProvisioningParametersDTO, error) {
	var ersContext internal.ERSContext
	var parameters internal.ProvisioningParametersDTO

	if details.ServiceID != KymaServiceID {
		return ersContext, parameters, errors.New("service_id not recognized")
	}
	if _, exists := b.enabledPlanIDs[details.PlanID]; !exists {
		return ersContext, parameters, errors.Errorf("plan ID %q is not recognized", details.PlanID)
	}

	result, err := b.plansSchemaValidator[details.PlanID].ValidateString(string(details.RawParameters))
	if err != nil {
		return ersContext, parameters, errors.Wrap(err, "while executing JSON schema validator")
	}

	if !result.Valid {
		return ersContext, parameters, errors.Wrapf(result.Error, "while validating input parameters")
	}

	ersContext, err = b.extractERSContext(details)
	if err != nil {
		return ersContext, parameters, errors.Wrap(err, "while extracting ers context")
	}

	parameters, err = b.extractInputParameters(details)
	if err != nil {
		return ersContext, parameters, errors.Wrap(err, "while extracting input parameters")
	}

	if !b.kymaVerOnDemand && parameters.KymaVersion != "" {
		logger.Infof("Kyma on demand functionality is disabled. Default Kyma version will be used instead %s", parameters.KymaVersion)
		parameters.KymaVersion = ""
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

func (b *ProvisionEndpoint) handleExistingOperation(operation *internal.ProvisioningOperation, input internal.ProvisioningParameters, log logrus.FieldLogger) (domain.ProvisionedServiceSpec, error) {
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		log.Errorf("cannot get provisioning parameters from exist operation", err)
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
	msg := fmt.Sprintf("provisioning operation with InstanceID %s already exist", operation.InstanceID)
	return domain.ProvisionedServiceSpec{}, apiresponses.NewFailureResponse(err, http.StatusConflict, msg)
}
