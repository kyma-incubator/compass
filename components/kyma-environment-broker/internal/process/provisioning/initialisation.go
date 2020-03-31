package provisioning

import (
	"fmt"
	"net/url"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// the time after which the operation is marked as expired
	CheckStatusTimeout = 3 * time.Hour
)

//go:generate mockery -name=DirectorClient -output=automock -outpkg=automock -case=underscore

type DirectorClient interface {
	GetConsoleURL(accountID, runtimeID string) (string, error)
}

type InitialisationStep struct {
	operationManager  *OperationManager
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
	directorClient    DirectorClient
	inputBuilder      input.CreatorForPlan
}

func NewInitialisationStep(os storage.Operations, is storage.Instances, pc provisioner.Client, dc DirectorClient, b input.CreatorForPlan) *InitialisationStep {
	return &InitialisationStep{
		operationManager:  NewOperationManager(os),
		instanceStorage:   is,
		provisionerClient: pc,
		directorClient:    dc,
		inputBuilder:      b,
	}
}

func (s *InitialisationStep) Name() string {
	return "Provision_Initialization"
}

func (s *InitialisationStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	_, err := s.instanceStorage.GetByID(operation.InstanceID)
	switch {
	case err == nil:
		log.Info("instance exist, check instance status")
		return s.checkRuntimeStatus(operation, log)
	case dberr.IsNotFound(err):
		log.Info("instance not exist, initialize runtime input request")
		return s.initializeRuntimeInputRequest(operation, log)
	default:
		log.Errorf("unable to get instance from storage: %s", err)
		return operation, 1 * time.Second, nil
	}
}

func (s *InitialisationStep) initializeRuntimeInputRequest(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		log.Errorf("cannot fetch provisioning parameters from operation: %s", err)
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	log.Infof("create input creator for %q plan ID", pp.PlanID)
	creator, found := s.inputBuilder.ForPlan(pp.PlanID)
	if !found {
		log.Error("input creator does not exist")
		return s.operationManager.OperationFailed(operation, "cannot create provisioning input creator")
	}

	operation.InputCreator = creator
	return operation, 0, nil
}

func (s *InitialisationStep) checkRuntimeStatus(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > CheckStatusTimeout {
		log.Infof("operation has reached the time limit: updated operation time: %s", operation.UpdatedAt)
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", CheckStatusTimeout))
	}

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	if err != nil {
		return operation, 10 * time.Second, nil
	}

	_, err = url.ParseRequestURI(instance.DashboardURL)
	if err == nil {
		return s.operationManager.OperationSucceeded(operation, "URL dashboard already exist")
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
		repeat, err := s.handleDashboardURL(instance)
		if err != nil || repeat != 0 {
			return operation, repeat, err
		}
		return s.operationManager.OperationSucceeded(operation, msg)
	case gqlschema.OperationStateInProgress:
		return operation, 2 * time.Minute, nil
	case gqlschema.OperationStatePending:
		return operation, 2 * time.Minute, nil
	case gqlschema.OperationStateFailed:
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("provisioner client returns failed status: %s", msg))
	}

	return s.operationManager.OperationFailed(operation, fmt.Sprintf("unsupported provisioner client status: %s", status.State.String()))
}

func (s *InitialisationStep) handleDashboardURL(instance *internal.Instance) (time.Duration, error) {
	dashboardURL, err := s.directorClient.GetConsoleURL(instance.GlobalAccountID, instance.RuntimeID)
	if director.IsTemporaryError(err) {
		return 3 * time.Minute, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "while geting URL from director")
	}

	instance.DashboardURL = dashboardURL
	err = s.instanceStorage.Update(*instance)
	if err != nil {
		return 10 * time.Second, nil
	}

	return 0, nil
}
