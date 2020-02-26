package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/cmd/poc-eventhubs-provisioner/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/cmd/poc-eventhubs-provisioner/k8s"
)

// Hardcoded values for the PoC
const (
	groupName             = "poc-sample-resource-group"
	eventHubsNamespace    = "poc-sample-eventhubs-namespace"
	authorizationRuleName = "RootManageSharedAccessKey"

	kafkaPort     = 9093
	kafkaUsername = "kafka-username"

	k8sSecretName      = "secret-name"
	k8sSecretNamespace = "secret-namespace"
)

func main() {
	cfg, err := azure.GetConfigFromEnvironment("test.env")
	if err != nil {
		log.Fatalf("Failed to get config from env: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	defer func() {
		if err := azure.Cleanup(context.Background(), cfg, groupName); err != nil {
			log.Fatalf("Failed to cleanup resources with error: %v", err)
		}
	}()

	if _, err = azure.PersistResourceGroup(ctx, cfg, groupName); err != nil {
		log.Fatalf("Failed to persist Azure Resource Group [%s] with error: %v", groupName, err)
	}
	log.Printf("Persisted Azure Resource Group [%s]", groupName)

	if _, err = azure.PersistEventHubsNamespace(ctx, cfg, groupName, eventHubsNamespace); err != nil {
		log.Fatalf("Failed to persist Azure EventHubs Namespace [%s] with error: %v", eventHubsNamespace, err)
	}
	log.Printf("Persisted Azure EventHubs Namespace [%s]", eventHubsNamespace)

	log.Printf("Get Access Keys for Azure EventHubs Namespace [%s]", eventHubsNamespace)
	accessKeys, err := azure.GetEventHubsNamespaceAccessKeys(ctx, cfg, groupName, eventHubsNamespace, authorizationRuleName)
	if err != nil {
		log.Fatalf("Failed to get Access Keys for Azure EventHubs Namespace [%s] with error: %v", eventHubsNamespace, err)
	}

	kafkaEndpoint := extractEndpoint(accessKeys)
	kafkaEndpoint = appendPort(kafkaEndpoint, kafkaPort)
	kafkaPassword := *accessKeys.PrimaryConnectionString

	secret := k8s.SecretFrom(k8sSecretName, k8sSecretNamespace, kafkaEndpoint, kafkaUsername, kafkaPassword, eventHubsNamespace)
	secretBytes, _ := json.MarshalIndent(secret, "", "    ")
	log.Printf("Kubernetes secret:\n%s", secretBytes)

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
