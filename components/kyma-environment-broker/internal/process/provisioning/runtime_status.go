package provisioning

import (
	"fmt"
	"net/url"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// the time after which the operation is marked as expired
	CheckStatusTimeout = 3 * time.Hour
)

type DirectorClient interface {
	GetConsoleURL(accountID, runtimeID string) (string, error)
}

type RuntimeStatusStep struct {
	operationManager  *process.OperationManager
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
	DirectorClient    DirectorClient
}

func NewRuntimeStatusStep(os storage.Operations, is storage.Instances, pc provisioner.Client, dc DirectorClient) *RuntimeStatusStep {
	return &RuntimeStatusStep{
		operationManager:  process.NewOperationManager(os),
		instanceStorage:   is,
		provisionerClient: pc,
		DirectorClient:    dc,
	}
}

func (s *RuntimeStatusStep) Name() string {
	return "Check_Runtime_Status"
}

func (s *RuntimeStatusStep) Run(operation internal.ProvisioningOperation, log *logrus.Entry) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > CheckStatusTimeout {
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", CheckStatusTimeout))
	}

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	if err != nil {
		return operation, 1 * time.Minute, nil
	}

	_, err = url.ParseRequestURI(instance.DashboardURL)
	if err == nil {
		return s.operationManager.OperationSucceeded(operation, "URL dashboard already exist")
	}

	status, err := s.provisionerClient.RuntimeOperationStatus(instance.GlobalAccountID, operation.ProvisionerOperationID)
	if err != nil {
		return operation, 1 * time.Minute, nil
	}
	log.Infof("Call to provisioner returned %s status", status.State.String())

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
		return operation, 10 * time.Minute, nil
	case gqlschema.OperationStatePending:
		return operation, 10 * time.Minute, nil
	case gqlschema.OperationStateFailed:
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("provisioner client returns failed status: %s", msg))
	}

	return s.operationManager.OperationFailed(operation, fmt.Sprintf("unsupported provisioner client status: %s", status.State.String()))
}

func (s *RuntimeStatusStep) handleDashboardURL(instance *internal.Instance) (time.Duration, error) {
	dashboardURL, err := s.DirectorClient.GetConsoleURL(instance.GlobalAccountID, instance.RuntimeID)
	if director.IsTemporaryError(err) {
		return 10 * time.Minute, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "while geting URL from director")
	}

	instance.DashboardURL = dashboardURL
	err = s.instanceStorage.Update(*instance)
	if err != nil {
		return 1 * time.Minute, nil
	}

	return 0, nil
}
