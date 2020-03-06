package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/k8s"
)

// Hardcoded values for the PoC
const (
	authorizationRuleName = "RootManageSharedAccessKey"

	kafkaPort = 9093

	k8sSecretName      = "secret-name"
	k8sSecretNamespace = "secret-namespace"
)

// This example code shows how an unused EventHub can be retrieved from a given Azure subscription
// the EventHub Namespace is marked as used after one run
// to rerun set the in_use tag of the Namespace to false
func main() {
	cfg, err := azure.GetConfigFromEnvironment("/Users/i512777/tickets/7242-poc-azure-eventhubs-namespace-provisioner/test.env")
	if err != nil {
		log.Fatalf("Failed to get config from env: %v", err)
	}

	ctx := context.Background()

	unusedEventHubNamespace, err := azure.GetFirstUnusedNamespaces(ctx, cfg)
	if err != nil {
		panic("no ready EventHubs Namespace found")
	}
	log.Printf("Get Access Keys for Azure EventHubs Namespace [%s]\n", unusedEventHubNamespace)
	resourceGroup := azure.GetResourceGroup(unusedEventHubNamespace)

	log.Printf("Found unused EventHubs Namespace, name: %v, resourceGroup: %v", unusedEventHubNamespace.Name, resourceGroup)

	accessKeys, err := azure.GetEventHubsNamespaceAccessKeys(ctx, cfg, resourceGroup, *unusedEventHubNamespace.Name, authorizationRuleName)
	if err != nil {
		log.Fatalf("Failed to get Access Keys for Azure EventHubs Namespace [%s] with error: %v\n", unusedEventHubNamespace, err)
	}

	kafkaEndpoint := extractEndpoint(accessKeys)
	kafkaEndpoint = appendPort(kafkaEndpoint, kafkaPort)
	kafkaPassword := *accessKeys.PrimaryConnectionString

	secret := k8s.SecretFrom(k8sSecretName, k8sSecretNamespace, kafkaEndpoint, kafkaPassword, *unusedEventHubNamespace.Name)
	secretBytes, _ := json.MarshalIndent(secret, "", "    ")
	log.Printf("Kubernetes secret:\n%s", secretBytes)

	if _, err := azure.MarkNamespaceAsUsed(ctx, cfg, resourceGroup, unusedEventHubNamespace); err != nil {
		panic(err)
	}

	// Note: At this point, the PoC is completed:
	// - Provision an Azure EventHubs Namespace inside a new or existing Azure Resource Group.
	// - Prepare a Kubernetes secret for the provisioned Azure EventHubs Namespace.

	// Note: The created secret should be persisted to the desired Kubernetes cluster
	// for the Eventing flow to work from that cluster to the provisioned Azure Event Hubs Namespace.
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
