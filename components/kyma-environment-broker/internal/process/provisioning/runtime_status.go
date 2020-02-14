package provisioning

import (
	"net/url"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

const (
	// time delay after which the instance becomes obsolete in the process of polling for last operation
	delayInstanceTime = 3 * time.Hour
)

type DirectorClient interface {
	GetConsoleURL(accountID, runtimeID string) (string, error)
}

type RuntimeStatusStep struct {
	operationStorage  storage.Operations
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
	DirectorClient    DirectorClient
}

func NewRuntimeStatusStep(os storage.Operations, is storage.Instances, pc provisioner.Client, dc DirectorClient) *RuntimeStatusStep {
	return &RuntimeStatusStep{
		operationStorage:  os,
		instanceStorage:   is,
		provisionerClient: pc,
		DirectorClient:    dc,
	}
}

func (s *RuntimeStatusStep) Run(operation *internal.ProvisioningOperation) (error, time.Duration) {
	// TODO check if should operation be processed
	// TODO mark operation as failed if error is permanent

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	if err != nil {
		return nil, 1 * time.Minute
	}
	_, err = url.ParseRequestURI(instance.DashboardURL)
	if err == nil {
		return nil, 0
	}

	status, err := s.provisionerClient.RuntimeOperationStatus(instance.GlobalAccountID, operation.ProvisionerOperationID)
	if err != nil {
		return nil, 1 * time.Minute
	}

	var msg string
	if status.Message != nil {
		msg = *status.Message
	}

	switch status.State {
	case gqlschema.OperationStateSucceeded:
		operation.State = domain.Succeeded
		operation.Description = msg
		return s.handleSuccessRuntime(operation, instance)
	case gqlschema.OperationStateInProgress:
		return nil, 10 * time.Minute
	case gqlschema.OperationStatePending:
		return nil, 10 * time.Minute
	case gqlschema.OperationStateFailed:
		return errors.Errorf("provisioner client returns failed status: %s", msg), 0
	}

	return errors.Errorf("unsupported provisioner client status: %s", status.State.String()), 0
}

func (s *RuntimeStatusStep) handleSuccessRuntime(operation *internal.ProvisioningOperation, instance *internal.Instance) (error, time.Duration) {
	err, repeat := s.handleDashboardURL(instance)
	if err != nil || repeat != 0 {
		return err, repeat
	}

	_, err = s.operationStorage.UpdateProvisioningOperation(*operation)
	if err != nil {
		return nil, 1 * time.Minute
	}

	return nil, 0
}

func (s *RuntimeStatusStep) handleDashboardURL(instance *internal.Instance) (error, time.Duration) {
	dashboardURL, err := s.DirectorClient.GetConsoleURL(instance.GlobalAccountID, instance.RuntimeID)
	if director.IsTemporaryError(err) {
		return s.checkInstanceOutdated(instance)
	}
	if err != nil {
		return errors.Wrapf(err, "while geting URL from director"), 0
	}

	instance.DashboardURL = dashboardURL
	err = s.instanceStorage.Update(*instance)
	if err != nil {
		return nil, 1 * time.Minute
	}

	return nil, 0
}

func (s *RuntimeStatusStep) checkInstanceOutdated(instance *internal.Instance) (error, time.Duration) {
	addTime := instance.CreatedAt.Add(delayInstanceTime)
	subTime := time.Now().Sub(addTime)

	if subTime > 0 {
		// after delayInstanceTime Instance last operation is marked as failed
		return errors.Errorf(""), 0
	}

	return nil, 10 * time.Minute
}
