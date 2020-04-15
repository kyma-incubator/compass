package azure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
)

const AzureFutureOperationSucceeded string = "Succeeded"

// TODO(nachtmaar): delete me later, doesn't make sense in CI without credentials
func TestDeleteResourceGroup(t *testing.T) {
	t.Skip("remove me if you wanna run this test")
	// given
	tags := Tags{TagInstanceID: ptr.String("12345678")}
	azureProvider := NewAzureProvider()
	azureConfig, err := GetConfigFromEnvironment()
	require.NoError(t, err)
	azureClient, err := azureProvider.GetClient(azureConfig, logrus.New())
	require.NoError(t, err)
	ctx := context.Background()

	// when
	for {
		time.Sleep(time.Second)

		resourceGroup, err := azureClient.GetResourceGroup(ctx, tags)
		if err != nil {
			// TODO: check for does not exist
			t.Logf("error in getting resource group: %deprovisioningState", err)
			if err, ok := err.(ResourceGroupDoesNotExist); ok {
				t.Log("resource group already exists, we are done here ...")
				t.Log(err)
				break
			}
			continue
		}
		deprovisioningState := *resourceGroup.Properties.ProvisioningState
		t.Logf("provisioning state: %s", deprovisioningState)
		if deprovisioningState != "Deleting" {
			future, err := azureClient.DeleteResourceGroup(ctx, tags)
			if _, ok := err.(*ResourceGroupDoesNotExist); ok {
				t.Log("resource group already exists, we are done here ...")
				break
			}
			if err != nil {
				errorMessage := fmt.Sprintf("Unable to delete Azure resource group: %v", err)
				t.Logf(errorMessage)
				continue
			}
			if future.Status() != AzureFutureOperationSucceeded {
				t.Logf("rescheduling step to check deletion of resource group completed")
				continue
			}
		}
	}
	// then
	// TODO:
}

func GetConfigFromEnvironment() (*Config, error) {
	err := envy.Load("/Users/i512777/tickets/7242-poc-azure-eventhubs-namespace-provisioner/test.env")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %s", err)
	}

	clientID, err := envy.MustGet("AZURE_CLIENT_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	clientSecret, err := envy.MustGet("AZURE_CLIENT_SECRET")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	tenantID, err := envy.MustGet("AZURE_TENANT_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	subscriptionID, err := envy.MustGet("AZURE_SUBSCRIPTION_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	azureConfig, err := GetConfig(clientID, clientSecret, tenantID, subscriptionID, "westeurope")
	return azureConfig, nil
}
