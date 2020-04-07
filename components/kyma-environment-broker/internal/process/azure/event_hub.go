package azure

import (
	"context"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
)

type EventHub struct {
	HyperscalerProvider azure.HyperscalerProvider
	AccountProvider     hyperscaler.AccountProvider
	Context             context.Context
}


//TODO(montaro) desperate attempt to consolidate the shared code between provision and deprovision steps
//func (eh EventHub)shared(s deprovisioning.Step, operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.OperationInterface, time.Duration, error) {
//
//	hypType := hyperscaler.Azure
//
//	//parse provisioning parameters
//	pp, err := operation.GetParameters()
//	if err != nil {
//		// if the parameters are incorrect, there is no reason to retry the operation
//		// a new request has to be issued by the user
//		log.Errorf("Aborting after failing to get valid operation provisioning parameters: %v", err)
//		return s.(deprovisioning.DeprovisionAzureEventHubStep).OperationManager.OperationFailed(operation, "invalid operation provisioning parameters")
//
//	}
//	log.Infof("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)
//
//	//get hyperscaler credentials from HAP
//	//credentials, err := s.(deprovisioning.DeprovisionAzureEventHubStep).EventHub.AccountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
//	credentials, err := eh.AccountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
//	if err != nil {
//		// retrying might solve the issue, the HAP could be temporarily unavailable
//		errorMessage := fmt.Sprintf("Unable to retrieve Gardener Credentials from HAP lookup: %v", err)
//		return s.(deprovisioning.DeprovisionAzureEventHubStep).OperationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
//	}
//	azureCfg, err := azure.GetConfigFromHAPCredentialsAndProvisioningParams(credentials, pp)
//	if err != nil {
//		// internal error, repeating doesn't solve the problem
//		errorMessage := fmt.Sprintf("Failed to create Azure config: %v", err)
//		return s.(deprovisioning.DeprovisionAzureEventHubStep).OperationManager.OperationFailed(operation, errorMessage)
//	}
//
//	// create hyperscaler client
//	//namespaceClient, err := s.(deprovisioning.DeprovisionAzureEventHubStep).HyperscalerProvider.GetClient(azureCfg)
//	namespaceClient, err := eh.HyperscalerProvider.GetClient(azureCfg)
//	if err != nil {
//		// internal error, repeating doesn't solve the problem
//		errorMessage := fmt.Sprintf("Failed to create Azure EventHubs client: %v", err)
//		return s.(deprovisioning.DeprovisionAzureEventHubStep).OperationManager.OperationFailed(operation, errorMessage)
//	}
//	//TODO (montaro) return the namespaceClient
//	println(namespaceClient)
//	return nil, 0, nil
//
//}