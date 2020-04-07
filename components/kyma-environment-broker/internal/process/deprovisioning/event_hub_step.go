package deprovisioning

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
)

type DeprovisionAzureEventHubStep struct {
	operationManager *process.DeprovisionOperationManager
	hyperscalerProvider azure.HyperscalerProvider
	instanceStorage  storage.Instances
	accountProvider  hyperscaler.AccountProvider
	context context.Context
}

func NewDeprovisionAzureEventHubStep(os storage.Operations, is storage.Instances,hyperscalerProvider azure.HyperscalerProvider, accountProvider hyperscaler.AccountProvider, ctx context.Context) DeprovisionAzureEventHubStep {
	return DeprovisionAzureEventHubStep{
		operationManager:  process.NewDeprovisionOperationManager(os),
		hyperscalerProvider: hyperscalerProvider,
		instanceStorage:     is,
		accountProvider:     accountProvider,
		context:             ctx,
	}
}

var _ Step = (*DeprovisionAzureEventHubStep)(nil)

func (s DeprovisionAzureEventHubStep) Name() string {
	return "Deprovision Azure Event Hubs"
}


func (s DeprovisionAzureEventHubStep) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	// TODO(nachtmaar): global timeout as in remove_runtime.go ?
	// TODO(nachtmaar): have shared code between provisioning/deprovisioning
	hypType := hyperscaler.Azure

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	switch {
	case err == nil:
	case dberr.IsNotFound(err):
		return s.operationManager.OperationSucceeded(operation, "instance already deprovisioned")
	default:
		log.Errorf("unable to get instance from storage: %s", err)
		return operation, 1 * time.Second, nil
	}

	if instance.RuntimeID == "" {
		log.Warn("Runtime not exist")
		return s.operationManager.OperationSucceeded(operation, "runtime was never provisioned")
	}

	// get hyperscaler credentials from HAP
	credentials, err := s.accountProvider.GardenerCredentials(hypType, instance.GlobalAccountID)
	if err != nil {
		// retrying might solve the issue, the HAP could be temporarily unavailable
		errorMessage := fmt.Sprintf("Unable to retrieve Gardener Credentials from HAP lookup: %v", err)
		return s.operationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	azureCfg, err := azure.GetConfigFromHAPCredentialsAndProvisioningParams(credentials, operation.GetOperation())
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure config: %v", err)
		return s.operationManager.OperationFailed(operation, errorMessage)
	}

	// create hyperscaler client
	namespaceClient, err := s.hyperscalerProvider.GetClient(azureCfg)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure EventHubs client: %v", err)
		return s.operationManager.OperationFailed(operation, errorMessage)
	}
	if err := namespaceClient.DeleteResourceGroup(s.context, instance.InstanceID); err != nil {
		// TODO(nachtmaar): be nice and create s.operationmanager.RetryForever
		return operation, 20 * time.Second, nil
	}

	// TODO(nachtmaar): be nice and create s.operationmanager.RetryForever
	return s.operationManager.OperationSucceeded(operation, "deproviosioning of event_hub_step succeeded")
}

