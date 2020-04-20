package provisioning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
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

	// Generating a unique name for resource group and namespace with operation ID as suffix if available
	uniqueName, err := p.getUniqueNSName(operation.ID, azureClient, log)
	if err != nil {
		return p.operationManager.OperationFailed(operation, fmt.Sprintf("Can't find a unique name for Event Hubs namespace. Error:%v", err))
	}

	// create Resource Group
	groupName := uniqueName
	// TODO(nachtmaar): use different resource group name https://github.com/kyma-incubator/compass/issues/967
	resourceGroup, err := azureClient.CreateResourceGroup(p.context, azureCfg, groupName, tags)
	if err != nil {
		// retrying might solve the issue while communicating with azure, e.g. network problems etc
		errorMessage := fmt.Sprintf("Failed to create Azure Resource Group [%s] with error: %v", groupName, err)
		return p.operationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	log.Info("created Azure Resource Group [%s]", groupName)

	// create EventHubs Namespace
	namespaceName := uniqueName
	ns, err := azureClient.CreateNamespace(p.context, azureCfg, groupName, namespaceName, tags)
	if err != nil {
		// retrying might solve the issue while communicating with azure, e.g. network problems etc
		errorMessage := fmt.Sprintf("Failed to create Azure Event Hubs namespace [%s] with error: %v", namespaceName, err)
		return p.operationManager.RetryOperation(operation, errorMessage, time.Minute, time.Minute*30, log)
	}
	log.Infof("created Azure Event Hubs namespace [%s]", namespaceName)

	// get EventHubs Namespace secret
	accessKeys, err := azureClient.GetEventhubAccessKeys(p.context, *resourceGroup.Name, *ns.Name, authorizationRuleName)
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

// getUniqueNSName will try to use suffix to generate a unique name for Event Hubs namespace, or generate a random one
// if it already exists
func (p *ProvisionAzureEventHubStep) getUniqueNSName(suffix string, azureClient azure.AzureInterface,
	log logrus.FieldLogger) (string, error) {
	uniqueName := fmt.Sprintf("skr-%s", suffix)
	u, err := azureClient.CheckNamespaceAvailability(p.context, uniqueName)
	if err != nil {
		log.Errorf("Error while calling Azure Event Hubs client. Can't check namespace name availability. "+
			"Error: %v", err)
		return "", errors.Errorf("Error while calling Azure Event Hubs client. Can't check namespace name "+
			"availability. Error: %v", err)
	}

	attempts := 10
	for !u && attempts > 0 {
		newName := fmt.Sprintf("skr-%s", uuid.New().String())
		log.Info("Namespace %s already exists. Checking new name %s.", uniqueName, newName)
		uniqueName = newName
		u, err = azureClient.CheckNamespaceAvailability(p.context, uniqueName)
		if err != nil {
			log.Errorf("Error while calling Azure Event Hubs client. Can't check namespace name availability. "+
				"Error: %v", err)
			return "", errors.Errorf("Error while calling Azure Event Hubs client. Can't check namespace name "+
				"availability. Error: %v", err)
		}
		attempts--
	}

	if !u {
		log.Errorf("Can't find a unique name for Event Hubs namespace. Failed after 10 attempts.")
		return "", errors.New("Can't find a unique name for Event Hubs namespace. Failed after 10 attempts.")
	}
	return uniqueName, nil
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
