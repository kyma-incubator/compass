package event_hub

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

const (
	authorizationRuleName = "RootManageSharedAccessKey"

	kafkaPort = 9093

	k8sSecretName                     = "secret-name"
	k8sSecretNamespace                = "knative-eventing"
	componentNameKnativeEventing      = "knative-eventing"
	componentNameKnativeEventingKafka = "knative-eventing-kafka"
	kafkaProvider                     = "azure"
)

/*TODO(anishj0shi)
- Implement retry logic for Namespace retrieval and NamespaceTagging operation.
*/

type ProvisionAzureEventHubStep struct {
	operationManager *process.OperationManager
	namespaceClient  azure.NamespaceClientInterface
	accountProvider  hyperscaler.AccountProvider
	context          context.Context
}

func NewProvisionAzureEventHubStep(os storage.Operations, namespaceClient azure.NamespaceClientInterface, accountProvider hyperscaler.AccountProvider, ctx context.Context) *ProvisionAzureEventHubStep {
	return &ProvisionAzureEventHubStep{
		operationManager: process.NewOperationManager(os),
		namespaceClient:  namespaceClient,
		accountProvider:  accountProvider,
		context:          ctx,
	}
}

func (p *ProvisionAzureEventHubStep) Name() string {
	return "Provision Azure Event Hubs"
}

func (p *ProvisionAzureEventHubStep) Run(operation internal.ProvisioningOperation,
	log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	pp, err := operation.GetProvisioningParameters()
	log.Infof("Provisioning Params..... %+v", pp) //TODO(anishj0shi) Remove after testing

	if err != nil {
		log.Error("Aborting after failing to get valid operation provisioning parameters")
		return p.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	hypType := hyperscaler.Azure

	log.Infof("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	credentials, err := p.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)

	if err != nil {
		log.Error("Unable to retrieve Gardener Credentials from HAP lookup")
		return operation, 5 * time.Second, nil
	}

	log.Infof("CREDENTIALS RETRIEVED..... %+v", credentials) //TODO(anishj0shi) Remove after testing

	azureCfg, err := azure.GetConfigfromHAPCredentialsAndProvisioningParams(credentials, pp)

	unusedEventHubNamespace, err := azure.GetFirstUnusedNamespaces(p.context, p.namespaceClient)
	if err != nil {
		return p.operationManager.OperationFailed(operation, "no azure event-hub namespace found in the given subscription")
	}
	log.Printf("Found unused EventHubs Namespace, name: %v", unusedEventHubNamespace.Name)

	log.Infof("Get Access Keys for Azure EventHubs Namespace [%s]\n", unusedEventHubNamespace)
	resourceGroup := azure.GetResourceGroup(unusedEventHubNamespace)

	log.Printf("Found the unused EventHubs Namespace %v in resourceGroup: %v", unusedEventHubNamespace.Name, resourceGroup)

	accessKeys, err := azure.GetEventHubsNamespaceAccessKeys(p.context, p.namespaceClient, resourceGroup, *unusedEventHubNamespace.Name, authorizationRuleName)
	if err != nil {
		return p.operationManager.OperationFailed(operation, "unable to retrieve access keys to azure event-hub namespace")
	}

	kafkaEndpoint := extractEndpoint(accessKeys)
	kafkaEndpoint = appendPort(kafkaEndpoint, kafkaPort)
	kafkaPassword := *accessKeys.PrimaryConnectionString

	if _, err := azure.MarkNamespaceAsUsed(p.context, p.namespaceClient, resourceGroup, unusedEventHubNamespace); err != nil {
		return p.operationManager.OperationFailed(operation, "no azure event-hub namespace found in the given subscription")
	}

	// TODO(nachtmaar):
	// kafkaEndpoint := "TODO"
	// kafkaPassword := "TODO"

	operation.InputCreator.SetOverrides(componentNameKnativeEventing, getKnativeEventingOverrides())
	operation.InputCreator.SetOverrides(componentNameKnativeEventingKafka, getKafkaChannelOverrides(kafkaEndpoint, k8sSecretNamespace, "$ConnectionString", kafkaPassword, kafkaProvider))

	return operation, 0, nil
}

func extractEndpoint(accessKeys *eventhub.AccessKeys) string {
	endpoint := strings.Split(*accessKeys.PrimaryConnectionString, ";")[0]
	endpoint = strings.TrimPrefix(endpoint, "Endpoint=sb://")
	endpoint = strings.TrimSuffix(endpoint, "/")
	return endpoint
}

func appendPort(endpoint string, port int64) string {
	return fmt.Sprintf("%s:%d", endpoint, port)
}

func getKnativeEventingOverrides() []*gqlschema.ConfigEntryInput {
	var knativeOverrides []*gqlschema.ConfigEntryInput
	knativeOverrides = []*gqlschema.ConfigEntryInput{
		{
			Key:    "knative-eventing.channel.default.apiVersion",
			Value:  "knativekafka.kyma-project.io/v1alpha1",
			Secret: ptr.Bool(false),
		},
		{
			Key:    "knative-eventing.channel.default.kind",
			Value:  "KafkaChannel",
			Secret: ptr.Bool(false),
		},
	}
	return knativeOverrides
}

func getKafkaChannelOverrides(broker, namespace, username, password, kafkaProvider string) []*gqlschema.ConfigEntryInput {
	kafkaOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:    "kafka.brokers",
			Value:  broker,
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
	return kafkaOverrides
}
