package provisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
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

type CreateRuntimeStep struct {
	operationManager  *process.ProvisionOperationManager
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
}

func NewCreateRuntimeStep(os storage.Operations, is storage.Instances, cli provisioner.Client) *CreateRuntimeStep {
	return &CreateRuntimeStep{
		operationManager:  process.NewProvisionOperationManager(os),
		instanceStorage:   is,
		provisionerClient: cli,
	}
}

func (s *CreateRuntimeStep) Name() string {
	return "Create_Runtime"
}

func (s *CreateRuntimeStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > CreateRuntimeTimeout {
		log.Infof("operation has reached the time limit: updated operation time: %s", operation.UpdatedAt)
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", CreateRuntimeTimeout))
	}

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	requestInput, err := s.createProvisionInput(operation, pp)
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation data - cannot create provisioning input")
	}

	var provisionerResponse gqlschema.OperationStatus
	if operation.ProvisionerOperationID == "" {
		provisionerResponse, err := s.provisionerClient.ProvisionRuntime(pp.ErsContext.GlobalAccountID, pp.ErsContext.SubAccountID, requestInput)
		if err != nil {
			log.Errorf("call to provisioner failed: %s", err)
			return operation, 5 * time.Second, nil
		}
		operation.ProvisionerOperationID = *provisionerResponse.ID
		operation, repeat := s.operationManager.UpdateOperation(operation)
		if repeat != 0 {
			log.Errorf("cannot save operation ID from provisioner")
			return operation, 5 * time.Second, nil
		}
	}

	if provisionerResponse.RuntimeID == nil {
		provisionerResponse, err = s.provisionerClient.RuntimeOperationStatus(pp.ErsContext.GlobalAccountID, operation.ProvisionerOperationID)
		if err != nil {
			log.Errorf("call to provisioner about operation status failed: %s", err)
			return operation, 1 * time.Minute, nil
		}
	}
	if provisionerResponse.RuntimeID == nil {
		return operation, 1 * time.Minute, nil
	}
	log.Infof("fetched RuntimeID=%s", *provisionerResponse.RuntimeID)

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	if err != nil {
		log.Errorf("cannot get instance: %s", err)
		return operation, 1 * time.Minute, nil
	}
	instance.RuntimeID = *provisionerResponse.RuntimeID

	err = s.instanceStorage.Update(*instance)
	if err != nil {
		log.Errorf("cannot update instance in storage: %s", err)
		return operation, 10 * time.Second, nil
	}

	log.Info("runtime creation process initiated successfully")
	// return repeat mode (1 sec) to start the initialization step which will now check the runtime status
	return operation, 1 * time.Second, nil
}

func (s *CreateRuntimeStep) createProvisionInput(operation internal.ProvisioningOperation, parameters internal.ProvisioningParameters) (gqlschema.ProvisionRuntimeInput, error) {
	var request gqlschema.ProvisionRuntimeInput

	operation.InputCreator.SetProvisioningParameters(parameters.Parameters)
	operation.InputCreator.SetRuntimeLabels(operation.InstanceID, parameters.ErsContext.SubAccountID)
	request, err := operation.InputCreator.Create()
	if err != nil {
		return request, errors.Wrap(err, "while building input for provisioner")
	}

	return request, nil
}
