package event_hub

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

func fixLogger() logrus.FieldLogger {
	return logrus.StandardLogger()
}

/// A fake client for Azure EventHubs Namespace handling
type FakeNamespaceClient struct {
	eventHubs []eventhub.EHNamespace
}

func (nc *FakeNamespaceClient) ListKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	return eventhub.AccessKeys{}, nil
}

func (nc *FakeNamespaceClient) Update(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
	return parameters, nil
}

func (nc *FakeNamespaceClient) ListComplete(ctx context.Context) (result eventhub.EHNamespaceListResultIterator, err error) {
	return eventhub.EHNamespaceListResultIterator{}, nil
}

// ensure the fake client is implementing the interface
var _ azure.NamespaceClientInterface = (*FakeNamespaceClient)(nil)

func Test_ProvisionSucceeded(t *testing.T) {
	// given
	memoryStorageOp := storage.NewMemoryStorage().Operations()
	fakeNamespaceClient := FakeNamespaceClient{}
	step := NewProvisionAzureEventHubStep(memoryStorageOp, &fakeNamespaceClient, context.Background())
	op := internal.ProvisioningOperation{}

	// when
	op, _, err := step.Run(op, fixLogger())
	require.NoError(t, err)
	// require.Zero(t, when)

	// then
	// TODO(nachtmaar): validate overrides are correct
}
