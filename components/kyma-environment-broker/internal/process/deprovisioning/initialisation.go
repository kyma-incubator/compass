package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"github.com/pivotal-cf/brokerapi/v7/domain"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/sirupsen/logrus"
)

const (
	// the time after which the operation is marked as expired
	CheckStatusTimeout = 3 * time.Hour
)

type InitialisationStep struct {
	operationManager  *process.DeprovisionOperationManager
	operationStorage  storage.Provisioning
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
}

func NewInitialisationStep(os storage.Operations, is storage.Instances, pc provisioner.Client) *InitialisationStep {
	return &InitialisationStep{
		operationManager:  process.NewDeprovisionOperationManager(os),
		operationStorage:  os,
		instanceStorage:   is,
		provisionerClient: pc,
	}
}

func (s *InitialisationStep) Name() string {
	return "Deprovision_Initialization"
}

func (s *InitialisationStep) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	// rewrite necessary data from ProvisioningOperation to operation internal.DeprovisioningOperation
	op, err := s.operationStorage.GetProvisioningOperationByInstanceID(operation.InstanceID)
	if err != nil {
		log.Errorf("while getting provisioning operation from storage")
		return operation, time.Second * 10, nil
	}
	if op.State == domain.InProgress {
		log.Info("waiting for provisioning operation to finish")
		return operation, time.Minute, nil
	}

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	switch {
	case err == nil:
		if operation.ProvisionerOperationID == "" {
			return operation, 0, nil
		}
		log.Info("instance being removed, check operation status")
		return s.checkRuntimeStatus(operation, instance, log)
	case dberr.IsNotFound(err):
		return s.operationManager.OperationSucceeded(operation, "instance already deprovisioned")
	default:
		log.Errorf("unable to get instance from storage: %s", err)
		return operation, 1 * time.Second, nil
	}
}

func (s *InitialisationStep) checkRuntimeStatus(operation internal.DeprovisioningOperation, instance *internal.Instance, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > CheckStatusTimeout {
		log.Infof("operation has reached the time limit: updated operation time: %s", operation.UpdatedAt)
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", CheckStatusTimeout))
	}

	status, err := s.provisionerClient.RuntimeOperationStatus(instance.GlobalAccountID, operation.ProvisionerOperationID)
	if err != nil {
		return operation, 1 * time.Minute, nil
	}
	log.Infof("call to provisioner returned %s status", status.State.String())

	var msg string
	if status.Message != nil {
		msg = *status.Message
	}

	switch status.State {
	case gqlschema.OperationStateSucceeded:
		repeat, err := s.removeInstance(instance.InstanceID)
		if err != nil || repeat != 0 {
			return operation, repeat, err
		}
		return s.operationManager.OperationSucceeded(operation, msg)
	case gqlschema.OperationStateInProgress:
		return operation, 1 * time.Minute, nil
	case gqlschema.OperationStatePending:
		return operation, 1 * time.Minute, nil
	case gqlschema.OperationStateFailed:
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("provisioner client returns failed status: %s", msg))
	}

	return s.operationManager.OperationFailed(operation, fmt.Sprintf("unsupported provisioner client status: %s", status.State.String()))
}

func (s *InitialisationStep) removeInstance(instanceID string) (time.Duration, error) {
	err := s.instanceStorage.Delete(instanceID)
	if err != nil {
		return 10 * time.Second, nil
	}

	return 0, nil
}
