package provisioning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	retryCount            = 10
	authorizationRuleName = "RootManageSharedAccessKey"

	kafkaPort = "9093"

	k8sSecretNamespace                = "knative-eventing"
	componentNameKnativeEventing      = "knative-eventing"
	componentNameKnativeEventingKafka = "knative-eventing-kafka"
	kafkaProvider                     = "azure"
)

// ensure the interface is implemented
var _ Step = (*ProvisionAzureEventHubStep)(nil)

type ProvisionAzureEventHubStep struct {
	operationManager    *process.ProvisionOperationManager
	hyperscalerProvider azure.HyperscalerProvider
	accountProvider     hyperscaler.AccountProvider
	context             context.Context
}

func NewProvisionAzureEventHubStep(os storage.Operations, hyperscalerProvider azure.HyperscalerProvider, accountProvider hyperscaler.AccountProvider, ctx context.Context) *ProvisionAzureEventHubStep {
	return &ProvisionAzureEventHubStep{
		operationManager:    process.NewProvisionOperationManager(os),
		accountProvider:     accountProvider,
		context:             ctx,
		hyperscalerProvider: hyperscalerProvider,
	}
}

func (p *ProvisionAzureEventHubStep) Name() string {
	return "Provision Azure Event Hubs"
}

func (p *ProvisionAzureEventHubStep) Run(operation internal.ProvisioningOperation,
	log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	hypType := hyperscaler.Azure
	// parse provisioning parameters
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		// if the parameters are incorrect, there is no reason to retry the operation
		// a new request has to be issued by the user
		log.Errorf("Aborting after failing to get valid operation provisioning parameters: %v", err)
		return p.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}
	log.Infof("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	// get hyperscaler credentials from HAP
	credentials, err := p.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	if err != nil {
		// retrying might solve the issue, the HAP could be temporarily unavailable
		errorMessage := fmt.Sprintf("Unable to retrieve Gardener Credentials from HAP lookup: %v", err)
		return p.operationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	azureCfg, err := azure.GetConfigFromHAPCredentialsAndProvisioningParams(credentials, pp)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure config: %v", err)
		return p.operationManager.OperationFailed(operation, errorMessage)
	}

	// create hyperscaler client
	azureClient, err := p.hyperscalerProvider.GetClient(azureCfg)
	if err != nil {
		// internal error, repeating doesn't solve the problem
		errorMessage := fmt.Sprintf("Failed to create Azure EventHubs client: %v", err)
		return p.operationManager.OperationFailed(operation, errorMessage)
	}

	// prepare azure tags
	tags := azure.Tags{
		azure.TagSubAccountID: &pp.ErsContext.SubAccountID,
		azure.TagInstanceID:   &operation.InstanceID,
		azure.TagOperationID:  &operation.ID,
	}
	resourcesName := operation.ID
	exists, err := p.resourcesExist(resourcesName, tags, azureClient)

	if err != nil {
		return p.operationManager.RetryOperation(operation, fmt.Sprintf("Ceching Event Hubs resources failed"+
			". Error:%v", err), time.Minute, time.Minute*30, log)
	}

	if !exists {
		resourcesName, err = p.createEHResources(azureClient, azureCfg, tags, log)
		if err != nil {
			return p.operationManager.RetryOperation(operation, fmt.Sprintf("Creation of Event Hubs resources failed"+
				". Error:%v", err), time.Minute, time.Minute*30, log)
		}
	}

	// get EventHubs Namespace secret
	accessKeys, err := azureClient.GetEventhubAccessKeys(p.context, resourcesName, resourcesName, authorizationRuleName)
	if err != nil {
		// retrying might solve the issue while communicating with azure, e.g. network problems etc
		errorMessage := fmt.Sprintf("Unable to retrieve access keys to azure event-hub namespace: %v", err)
		return p.operationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	if accessKeys.PrimaryConnectionString == nil {
		// if GetEventhubAccessKeys() does not fail then a non-nil accessKey is returned
		// then retry the operation once
		errorMessage := "PrimaryConnectionString is nil"
		return p.operationManager.RetryOperationOnce(operation, errorMessage, time.Second*15, log)
	}
	kafkaEndpoint := extractEndpoint(accessKeys)
	kafkaPassword := *accessKeys.PrimaryConnectionString

	// set installation overrides
	operation.InputCreator.SetOverrides(componentNameKnativeEventing, getKnativeEventingOverrides())
	operation.InputCreator.SetOverrides(componentNameKnativeEventingKafka, getKafkaChannelOverrides(kafkaEndpoint, kafkaPort, k8sSecretNamespace, "$ConnectionString", kafkaPassword, kafkaProvider))

	return operation, 0, nil
}

func (p *ProvisionAzureEventHubStep) resourcesExist( resourcesName string, tags azure.Tags, azureClient azure.AzureInterface) (bool, error) {
	exists, err := azureClient.ResourceGroupExists(p.context, resourcesName, tags)
	if err != nil {
		return false, errors.Wrapf(err,"Can't check Azure resource group %s.", resourcesName)
	}
	// we are using the same name for resource group and namespace name
	exists, err = azureClient.NamespaceExists(p.context, resourcesName, resourcesName, tags)
	if err != nil {
		return false,errors.Wrapf(err, "Can't check Azure Event Hubs namespace %s.", resourcesName)
	}
	return exists, nil
}

func (p *ProvisionAzureEventHubStep) createEHResources( azureClient azure.AzureInterface, azureCfg *azure.Config, tags azure.Tags, log logrus.FieldLogger) (string, error) {
	uniqueName, err := p.getUniqueNSName(azureClient, log)
	if err != nil {
		return "", errors.Errorf("Can't find a unique name for Event "+
			"Hubs namespace. Error:%v", err)
	}

	// create Resource Group
	groupName := uniqueName
	if _, err = azureClient.CreateResourceGroup(p.context, azureCfg, groupName, tags); err != nil {
		// retrying might solve the issue while communicating with azure, e.g. network problems etc
		return "", errors.Errorf( "Failed to create Azure Resource Group [%s] with error: %v", groupName, err)
	}
	log.Info("created Azure Resource Group [%s]", groupName)

	// create EventHubs Namespace
	namespaceName := uniqueName
	if _, err = azureClient.CreateNamespace(p.context, azureCfg, groupName, namespaceName, tags); err != nil {
		return "", errors.Errorf("Failed to create Azure Event Hubs namespace [%s] with error: %v", namespaceName, err)
	}
	log.Infof("created Azure Event Hubs namespace [%s]", namespaceName)
	return uniqueName, nil
}

// getUniqueNSName will try to generate a unique name for Event Hubs namespace
func (p *ProvisionAzureEventHubStep) getUniqueNSName(azureClient azure.AzureInterface,
	log logrus.FieldLogger) (string, error) {
	for i := retryCount; i > 0; i-- {
		uniqueName := fmt.Sprintf("skr-%s", uuid.New().String())
		log.Info(" Checking name %s", uniqueName)
		var err error
		available, err := azureClient.CheckNamespaceAvailability(p.context, uniqueName)
		if err != nil {
			msg := "Error while calling Azure Event Hubs client. Can't check namespace name availability. Error: %v"
			log.Errorf(msg, err)
			return "", errors.Errorf(msg, err)
		}
		if available {
			return uniqueName, nil
		}
		log.Info("Namespace %s already exists.", uniqueName)
	}
	msg := fmt.Sprintf("Can't find a unique name for Event Hubs namespace. Failed after %d attempts.",
		retryCount)
	log.Error(msg)
	return "", errors.New(msg)
}

func extractEndpoint(accessKeys eventhub.AccessKeys) string {
	endpoint := strings.Split(*accessKeys.PrimaryConnectionString, ";")[0]
	endpoint = strings.TrimPrefix(endpoint, "Endpoint=sb://")
	endpoint = strings.TrimSuffix(endpoint, "/")
	return endpoint
}

func getKnativeEventingOverrides() []*gqlschema.ConfigEntryInput {
	return []*gqlschema.ConfigEntryInput{
		{
			Key:   "knative-eventing.channel.default.apiVersion",
			Value: "knativekafka.kyma-project.io/v1alpha1",
		},
		{
			Key:   "knative-eventing.channel.default.kind",
			Value: "KafkaChannel",
		},
	}
}

func getKafkaChannelOverrides(brokerHostname, brokerPort, namespace, username, password, kafkaProvider string) []*gqlschema.ConfigEntryInput {
	return []*gqlschema.ConfigEntryInput{
		{
			Key:    "kafka.brokers.hostname",
			Value:  brokerHostname,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.brokers.port",
			Value:  brokerPort,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.namespace",
			Value:  namespace,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.password",
			Value:  password,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.username",
			Value:  username,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.secretName",
			Value:  "knative-kafka",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "environment.kafkaProvider",
			Value:  kafkaProvider,
			Secret: ptr.Bool(true),
		},
	}
}
