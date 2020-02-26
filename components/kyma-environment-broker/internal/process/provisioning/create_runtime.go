package provisioning

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// the time after which the operation is marked as expired
	CreateRuntimeTimeout = 1 * time.Hour
)

type CreateRuntimeStep struct {
	operationManager  *process.OperationManager
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
	serviceManager    internal.ServiceManagerOverride
}

func NewCreateRuntimeStep(os storage.Operations, is storage.Instances, cli provisioner.Client, smOverride internal.ServiceManagerOverride) *CreateRuntimeStep {
	return &CreateRuntimeStep{
		operationManager:  process.NewOperationManager(os),
		instanceStorage:   is,
		provisionerClient: cli,
		serviceManager:    smOverride,
	}
}

func (s *CreateRuntimeStep) Name() string {
	return "Create_Runtime"
}

func (s *CreateRuntimeStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > CreateRuntimeTimeout {
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

	requestInput, err := s.createProvisionInput(operation, pp)
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation data - cannot create provisioning input")
	}

	var provisionerResponse gqlschema.OperationStatus
	if operation.ProvisionerOperationID == "" {
		provisionerResponse, err := s.provisionerClient.ProvisionRuntime(pp.ErsContext.GlobalAccountID, requestInput)
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
	if provisionerResponse.RuntimeID == nil {
		return operation, 5 * time.Minute, nil
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

func (s *CreateRuntimeStep) createProvisionInput(operation internal.ProvisioningOperation, parameters internal.ProvisioningParameters) (gqlschema.ProvisionRuntimeInput, error) {
	var request gqlschema.ProvisionRuntimeInput

	operation.InputCreator.SetOverrides(input.ServiceManagerComponentName, s.serviceManagerOverride(parameters.ErsContext))
	operation.InputCreator.SetProvisioningParameters(parameters.Parameters)
	operation.InputCreator.SetRuntimeLabels(operation.InstanceID, parameters.ErsContext.SubAccountID)
	request, err := operation.InputCreator.Create()
	if err != nil {
		return request, errors.Wrap(err, "while building input for provisioner")
	}

	return request, nil
}

func (s *CreateRuntimeStep) serviceManagerOverride(ersCtx internal.ERSContext) []*gqlschema.ConfigEntryInput {
	var smOverrides []*gqlschema.ConfigEntryInput
	if s.serviceManager.CredentialsOverride {
		smOverrides = []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: s.serviceManager.URL,
			},
			{
				Key:   "sm.user",
				Value: s.serviceManager.Username,
			},
			{
				Key:    "sm.password",
				Value:  s.serviceManager.Password,
				Secret: ptr.Bool(true),
			},
		}
	} else {
		smOverrides = []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: ersCtx.ServiceManager.URL,
			},
			{
				Key:   "sm.user",
				Value: ersCtx.ServiceManager.Credentials.BasicAuth.Username,
			},
			{
				Key:    "sm.password",
				Value:  ersCtx.ServiceManager.Credentials.BasicAuth.Password,
				Secret: ptr.Bool(true),
			},
		}
	}

	return smOverrides
}
