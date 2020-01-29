package broker

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

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
