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
	processazure "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)


type DeprovisionAzureEventHubStep struct {
	OperationManager *process.DeprovisionOperationManager
	processazure.EventHub
}

func NewDeprovisionAzureEventHubStep(os storage.Operations, hyperscalerProvider azure.HyperscalerProvider, accountProvider hyperscaler.AccountProvider, ctx context.Context) DeprovisionAzureEventHubStep {
	return DeprovisionAzureEventHubStep{
		OperationManager: process.NewDeprovisionOperationManager(os),
		EventHub: processazure.EventHub{
			HyperscalerProvider: hyperscalerProvider,
			AccountProvider:     accountProvider,
			Context:             ctx,
		},
	}
}

var _ Step = (*DeprovisionAzureEventHubStep)(nil)

func (s DeprovisionAzureEventHubStep) Name() string {
	return "Deprovision Azure Event Hubs"
}

func (s DeprovisionAzureEventHubStep) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	// TODO(nachtmaar): have shared code between provisioning/deprovisioning
	hypType := hyperscaler.Azure

	//parse provisioning parameters
	pp, err := operation.GetParameters()
	if err != nil {
		// if the parameters are incorrect, there is no reason to retry the operation
		// a new request has to be issued by the user
		log.Errorf("Aborting after failing to get valid operation provisioning parameters: %v", err)
		return s.OperationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}
	log.Infof("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	//get hyperscaler credentials from HAP
	credentials, err := s.EventHub.AccountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	if err != nil {
		// retrying might solve the issue, the HAP could be temporarily unavailable
		errorMessage := fmt.Sprintf("Unable to retrieve Gardener Credentials from HAP lookup: %v", err)
		return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	azureCfg, err := azure.GetConfigFromHAPCredentialsAndProvisioningParams(credentials, pp)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure config: %v", err)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}

	// create hyperscaler client
	namespaceClient, err := s.HyperscalerProvider.GetClient(azureCfg, log)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure EventHubs client: %v", err)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}
	// prepare azure tags
	tags := azure.Tags{azure.TagInstanceID: &operation.InstanceID}

	// check if resource group exists
	resourceGroup, err := namespaceClient.GetResourceGroup(s.EventHub.Context, tags)
	if err != nil {
		// if it doesn't exist anymore, there is nothing to delete - we are done
		if _, ok := err.(azure.ResourceGroupDoesNotExist); ok {
			return s.OperationManager.OperationSucceeded(operation, "deprovisioning of event_hub_step succeeded")
		}
		// custom error occurred while getting resource group - try again
		errorMessage := fmt.Sprintf("error while getting resource group, error: %s", err)
		return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
	}
	// delete the resource group if it still exists and deletion has not been triggered yet
	// TODO: check pointer
	deprovisioningState := *resourceGroup.Properties.ProvisioningState
	if deprovisioningState != azure.AzureFutureOperationDeleting {
		future, err := namespaceClient.DeleteResourceGroup(s.EventHub.Context, tags)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to delete Azure resource group: %v", err)
			return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
		}
		if future.Status() != azure.AzureFutureOperationSucceeded {
			//TODO(montaro) Optimization, retry after the polling delay provided by Azure
			if retryAfter, retryAfterHeaderExist := future.GetPollingDelay(); retryAfterHeaderExist {
				log.Infof("rescheduling step to check deletion of resource group completed")
				return s.OperationManager.RetryOperation(operation, "waiting for deprovisioning of azure resource group", retryAfter, time.Hour, log)
			} else {
				log.Infof("rescheduling step to check deletion of resource group completed")
				return s.OperationManager.RetryOperation(operation, "waiting for deprovisioning of azure resource group", time.Minute, time.Hour, log)
			}
		}
	}
	return s.OperationManager.RetryOperation(operation, "waiting for deprovisioning of azure resource group", time.Minute, time.Hour, log)
}
