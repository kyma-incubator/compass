package testing

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
)

// ensure the fake client is implementing the interface
var _ azure.AzureInterface = (*FakeNamespaceClient)(nil)

/// A fake client for Azure EventHubs Namespace handling
type FakeNamespaceClient struct {
	PersistEventhubsNamespaceError error
	ResourceGroupError             error
	AccessKeysError                error
	AccessKeys                     *eventhub.AccessKeys
	Tags                           azure.Tags
}

func (nc *FakeNamespaceClient) GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	if nc.AccessKeys != nil {
		return *nc.AccessKeys, nil
	}
	return eventhub.AccessKeys{
		PrimaryConnectionString: ptr.String("Endpoint=sb://name/;"),
	}, nc.AccessKeysError
}

func (nc *FakeNamespaceClient) CreateResourceGroup(ctx context.Context, config *azure.Config, name string, tags azure.Tags) (resources.Group, error) {
	nc.Tags = tags
	return resources.Group{
		Name: ptr.String("my-resourcegroup"),
	}, nc.ResourceGroupError
}

func (nc *FakeNamespaceClient) GetResourceGroup(ctx context.Context, tags azure.Tags) (resources.Group, error) {
	return resources.Group{}, nil
}

func (nc *FakeNamespaceClient) CreateNamespace(ctx context.Context, azureCfg *azure.Config, groupName, namespace string, tags azure.Tags) (*eventhub.EHNamespace, error) {
	nc.Tags = tags
	return &eventhub.EHNamespace{
		Name: ptr.String(namespace),
	}, nc.PersistEventhubsNamespaceError
}

func (nc *FakeNamespaceClient) DeleteResourceGroup(ctx context.Context, tags azure.Tags) (resources.GroupsDeleteFuture, error) {
	//TODO(montaro) double check me
	return resources.GroupsDeleteFuture{}, nil
}

func NewFakeNamespaceClientCreationError() azure.AzureInterface {
	return &FakeNamespaceClient{PersistEventhubsNamespaceError: fmt.Errorf("error while creating namespace")}
}

func NewFakeNamespaceClientListError() azure.AzureInterface {
	return &FakeNamespaceClient{AccessKeysError: fmt.Errorf("cannot list namespaces")}
}

func NewFakeNamespaceResourceGroupError() azure.AzureInterface {
	return &FakeNamespaceClient{ResourceGroupError: fmt.Errorf("cannot create resource group")}
}

func NewFakeNamespaceAccessKeysNil() azure.AzureInterface {
	return &FakeNamespaceClient{
		// no error here
		AccessKeysError: nil,
		AccessKeys: &eventhub.AccessKeys{
			// ups .. we got an AccessKeys with nil connection string even though there was no error
			PrimaryConnectionString: nil,
		},
	}
}

func NewFakeNamespaceClientHappyPath() *FakeNamespaceClient {
	return &FakeNamespaceClient{}
}

