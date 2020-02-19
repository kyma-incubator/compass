package provisioning

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// the time after which the operation is marked as expired
	CreateRuntimeTimeout = 1 * time.Hour
)

// Config holds all configurations connected with Provisioner API
type Config struct {
	URL             string
	GCPSecretName   string
	AzureSecretName string
	AWSSecretName   string
}

type CreateRuntimeStep struct {
	operationManager  *process.OperationManager
	instanceStorage   storage.Instances
	builderFactory    InputBuilderForPlan
	provisioningCfg   Config
	provisionerClient provisioner.Client
}

func NewCreateRuntimeStep(os storage.Operations, is storage.Instances, builderFactory InputBuilderForPlan, cfg Config, cli provisioner.Client) *CreateRuntimeStep {
	return &CreateRuntimeStep{
		operationManager:  process.NewOperationManager(os),
		instanceStorage:   is,
		builderFactory:    builderFactory,
		provisioningCfg:   cfg,
		provisionerClient: cli,
	}
}

func (s *CreateRuntimeStep) Name() string {
	return "Create_Runtime"
}

func (s *CreateRuntimeStep) Run(operation internal.ProvisioningOperation, log *logrus.Entry) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.CreatedAt) > CreateRuntimeTimeout {
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", CreateRuntimeTimeout))
	}

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}
	rawParameters, err := json.Marshal(pp.Parameters)
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation parameters")
	}

	input, err := s.createProvisionInput(pp)
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation data - cannot create provisioning input")
	}

	var provisionerResponse gqlschema.OperationStatus
	if operation.ProvisionerOperationID == "" {
		provisionerResponse, err := s.provisionerClient.ProvisionRuntime(pp.ErsContext.GlobalAccountID, input)
		if err != nil {
			log.Errorf("call to provisioner failed: %s", err)
			return operation, 1 * time.Minute, nil
		}
		operation.ProvisionerOperationID = *provisionerResponse.ID
		operation, repeat := s.operationManager.UpdateOperation(operation)
		if repeat != 0 {
			log.Errorf("cannot save operation ID from provisioner")
			return operation, 1 * time.Minute, nil
		}
	}

	if provisionerResponse.RuntimeID == nil {
		provisionerResponse, err = s.provisionerClient.RuntimeOperationStatus(pp.ErsContext.GlobalAccountID, operation.ProvisionerOperationID)
		if err != nil {
			log.Errorf("call to provisioner about operation status failed: %s", err)
			return operation, 5 * time.Minute, nil
		}
	}

	err = s.instanceStorage.Insert(internal.Instance{
		InstanceID:             operation.InstanceID,
		GlobalAccountID:        pp.ErsContext.GlobalAccountID,
		RuntimeID:              *provisionerResponse.RuntimeID,
		ServiceID:              pp.ServiceID,
		ServicePlanID:          pp.PlanID,
		ProvisioningParameters: string(rawParameters),
	})
	if err != nil {
		log.Errorf("cannot save instance in storage: %s", err)
		return operation, 1 * time.Minute, nil
	}

	return operation, 0, nil
}

func (s *CreateRuntimeStep) createProvisionInput(parameters internal.ProvisioningParameters) (gqlschema.ProvisionRuntimeInput, error) {
	var input gqlschema.ProvisionRuntimeInput

	inputBuilder, found := s.builderFactory.ForPlan(parameters.PlanID)
	if !found {
		return input, errors.Errorf("while finding input builder: plan %q not found.", parameters.PlanID)
	}
	inputBuilder.
		SetERSContext(parameters.ErsContext).
		SetProvisioningParameters(parameters.Parameters).
		SetProvisioningConfig(s.provisioningCfg)

	input, err := inputBuilder.Build()
	if err != nil {
		return input, errors.Wrap(err, "while building input for provisioner")
	}

	return input, nil
}
