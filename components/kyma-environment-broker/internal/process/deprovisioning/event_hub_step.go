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
	operationStorage storage.Operations
	instanceStorage  storage.Instances
	processazure.EventHub
}

func NewDeprovisionAzureEventHubStep(os storage.Operations, is storage.Instances, hyperscalerProvider azure.HyperscalerProvider, accountProvider hyperscaler.AccountProvider, ctx context.Context) DeprovisionAzureEventHubStep {
	return DeprovisionAzureEventHubStep{
		OperationManager: process.NewDeprovisionOperationManager(os),
		operationStorage: os,
		instanceStorage:  is,
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
	if operation.EventHub.Deleted {
		log.Info("Event Hub is already deprovisioned")
		return operation, 0, nil
	}

	hypType := hyperscaler.Azure

	instance, err := getInstance(s.instanceStorage, operation, log)
	switch err.(type) {
	case nil:
	case instanceNotFoundError:
		operation.EventHub.Deleted = true
		updatedOperation, err := s.operationStorage.UpdateDeprovisioningOperation(operation)
		if err != nil {
			log.Errorf("while updating the database for instance %s deprovision operation, error: %v",
				operation.InstanceID, err)
			return s.OperationManager.OperationSucceeded(operation, "instance already deprovisioned")
		} else {
			return s.OperationManager.OperationSucceeded(*updatedOperation, "instance already deprovisioned")
		}
	default:
		return operation, 1 * time.Second, nil
	}

	// parse provisioning parameters (instance being nil is protected by the previous switch case)
	pp, err := instance.GetProvisioningParameters()
	if err != nil {
		// if the parameters are incorrect, there is no reason to retry the operation
		// a new request has to be issued by the user
		errorMessage := fmt.Sprintf("aborting deprovisioning after failing to get valid operation provisioning parameters: %v", err)
		log.Errorf(errorMessage)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}
	log.Infof("HAP lookup for credentials to deprovision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	//get hyperscaler credentials from HAP
	credentials, err := s.EventHub.AccountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	if err != nil {
		// retrying might solve the issue, the HAP could be temporarily unavailable
		errorMessage := fmt.Sprintf("unable to retrieve Gardener Credentials from HAP lookup: %v", err)
		return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	azureCfg, err := azure.GetConfigFromHAPCredentialsAndProvisioningParams(credentials, pp)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("failed to create Azure config: %v", err)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}

	// create hyperscaler client
	namespaceClient, err := s.HyperscalerProvider.GetClient(azureCfg, log)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("failed to create Azure EventHubs client: %v", err)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}
	// prepare azure tags
	tags := azure.Tags{azure.TagInstanceID: &operation.InstanceID}

	// check if resource group exists
	resourceGroup, err := namespaceClient.GetResourceGroup(s.EventHub.Context, tags)
	if err != nil {
		// if it doesn't exist anymore, there is nothing to delete - we are done
		if _, ok := err.(azure.ResourceGroupDoesNotExistError); ok {
			if &resourceGroup != nil && resourceGroup.Name != nil {
				log.Infof("deprovisioning of event hub step succeeded, resource group: %v", resourceGroup.Name)
			} else {
				log.Info("deprovisioning of event hub step succeeded")
			}
			operation.EventHub.Deleted = true
			updatedOperation, err := s.operationStorage.UpdateDeprovisioningOperation(operation)
			if err != nil {
				log.Errorf("while updating the database, error: %v", err)
			}
			return *updatedOperation, 0, nil
		}
		// custom error occurred while getting resource group - try again
		errorMessage := fmt.Sprintf("error while getting resource group, error: %v", err)
		return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
	}
	// delete the resource group if it still exists and deletion has not been triggered yet
	if resourceGroup.Properties == nil || resourceGroup.Properties.ProvisioningState == nil {
		errorMessage := fmt.Sprintf("nil pointer in the resource group")
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}
	deprovisioningState := *resourceGroup.Properties.ProvisioningState
	if deprovisioningState != azure.FutureOperationDeleting {
		future, err := namespaceClient.DeleteResourceGroup(s.EventHub.Context, tags)
		if err != nil {
			errorMessage := fmt.Sprintf("unable to delete Azure resource group: %v", err)
			return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
		}
		if future.Status() != azure.FutureOperationSucceeded {
			var retryAfterDuration time.Duration
			if retryAfter, retryAfterHeaderExist := future.GetPollingDelay(); retryAfterHeaderExist {
				retryAfterDuration = retryAfter
			} else {
				retryAfterDuration = time.Minute
			}
			log.Info("rescheduling step to check deletion of resource group completed")
			return s.OperationManager.RetryOperation(operation, "waiting for deprovisioning of azure resource group", retryAfterDuration, time.Hour, log)
		}
	}
	errorMessage := "waiting for deprovisioning of azure resource group"
	return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
}
