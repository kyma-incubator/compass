package azure

import (
	"context"
	"fmt"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
)

// TODO(nachtmaar): delete me later, doesn't make sense in CI without credentials
func TestDeleteResourceGroup(t *testing.T) {
	t.Skip("remove me if you wanna run this test")
	// given
	tags := Tags{TagInstanceID: ptr.String("1234")}
	azureProvider := NewAzureProvider()
	azureConfig, err := GetConfigFromEnvironment()
	require.NoError(t, err)
	azureClient, err := azureProvider.GetClient(azureConfig)
	require.NoError(t, err)

	// when
	err = azureClient.DeleteResourceGroup(context.Background(), tags)
	require.NoError(t, err)

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
