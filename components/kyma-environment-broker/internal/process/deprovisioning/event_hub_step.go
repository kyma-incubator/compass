package deprovisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	processazure "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
)

type DeprovisionAzureEventHubStep struct {
	OperationManager *process.DeprovisionOperationManager
	instanceStorage  storage.Instances
	processazure.EventHub
}

func NewDeprovisionAzureEventHubStep(os storage.Operations, is storage.Instances, hyperscalerProvider azure.HyperscalerProvider, accountProvider hyperscaler.AccountProvider, ctx context.Context) DeprovisionAzureEventHubStep {
	return DeprovisionAzureEventHubStep{
		OperationManager: process.NewDeprovisionOperationManager(os),
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

// TODO(nachtmaar): move to better place
func GetParameters(provisioningParameters string) (internal.ProvisioningParameters, error) {
	var pp internal.ProvisioningParameters

	err := json.Unmarshal([]byte(provisioningParameters), &pp)
	if err != nil {
		return pp, errors.Wrap(err, "while unmarshaling provisioning parameters")
	}

	return pp, nil
}

func (s DeprovisionAzureEventHubStep) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	// TODO(nachtmaar): have shared code between provisioning/deprovisioning
	hypType := hyperscaler.Azure

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	switch {
	case err == nil:
	case dberr.IsNotFound(err):
		return s.OperationManager.OperationSucceeded(operation, "instance already deprovisioned")
	default:
		log.Errorf("unable to get instance from storage: %s", err)
		return operation, 1 * time.Second, nil
	}

	// parse provisioning parameters
	fmt.Printf("instance provisiong parameters :%s\n", instance.ProvisioningParameters)
	pp, err := GetParameters(instance.ProvisioningParameters)
	if err != nil {
		// if the parameters are incorrect, there is no reason to retry the operation
		// a new request has to be issued by the user
		log.Errorf("Aborting after failing to get valid operation provisioning parameters: %v", err)
		//TODO(montaro) Aborting the operation
		errorMessage := "invalid operation provisioning parameters"
		//return s.OperationManager.RetryForever(operation, "invalid operation provisioning parameters", time.Minute, log)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}
	log.Infof("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	//get hyperscaler credentials from HAP
	credentials, err := s.EventHub.AccountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	if err != nil {
		// retrying might solve the issue, the HAP could be temporarily unavailable
		errorMessage := fmt.Sprintf("Unable to retrieve Gardener Credentials from HAP lookup: %v", err)
		return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
		// return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	azureCfg, err := azure.GetConfigFromHAPCredentialsAndProvisioningParams(credentials, pp)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure config: %v", err)
		//TODO(montaro) Failing the operation assuming this is the desired behaviour
		//return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
		return s.OperationManager.OperationFailed(operation, errorMessage)
	}

	// create hyperscaler client
	namespaceClient, err := s.HyperscalerProvider.GetClient(azureCfg, log)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure EventHubs client: %v", err)
		//TODO(montaro) Failing the operation assuming this is the desired behaviour
		//return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
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
		return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
		// return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
	}
	// delete the resource group if it still exists and deletion has not been triggered yet
	// TODO: check pointer
	deprovisioningState := *resourceGroup.Properties.ProvisioningState
	if deprovisioningState != azure.AzureFutureOperationDeleting {
		future, err := namespaceClient.DeleteResourceGroup(s.EventHub.Context, tags)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to delete Azure resource group: %v", err)
			// return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
			return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
		}
		if future.Status() != azure.AzureFutureOperationSucceeded {
			//TODO(montaro) Optimization, retry after the polling delay provided by Azure
			if retryAfter, retryAfterHeaderExist := future.GetPollingDelay(); retryAfterHeaderExist {
				errorMessage := "rescheduling step to check deletion of resource group completed"
				log.Info(errorMessage)
				return s.OperationManager.RetryForever(operation, errorMessage, retryAfter, log)
				// return s.OperationManager.RetryOperation(operation, "waiting for deprovisioning of azure resource group", retryAfter, time.Hour, log)
			} else {
				errorMessage := "rescheduling step to check deletion of resource group completed"
				log.Infof(errorMessage)
				return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
				// return s.OperationManager.RetryOperation(operation, "waiting for deprovisioning of azure resource group", time.Minute, time.Hour, log)
			}
		}
	}
	errorMessage := "waiting for deprovisioning of azure resource group"
	return s.OperationManager.RetryForever(operation, errorMessage, time.Minute, log)
	// return s.OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Hour, log)
}
