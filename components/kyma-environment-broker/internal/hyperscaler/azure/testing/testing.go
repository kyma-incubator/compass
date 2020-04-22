package testing

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
)

// ensure the fake Client is implementing the interface
var _ azure.AzureInterface = (*FakeNamespaceClient)(nil)

/// A fake Client for Azure EventHubs Namespace handling
type FakeNamespaceClient struct {
	PersistEventhubsNamespaceError error
	ResourceGroupError             error
	AccessKeysError                error
	AccessKeys                     *eventhub.AccessKeys
	Tags                           azure.Tags
	GetResourceGroupError          error
	GetResourceGroupReturnValue    resources.Group
	DeleteResourceGroupCalled      bool
	DeleteResourceGroupError       error
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
	return nc.GetResourceGroupReturnValue, nc.GetResourceGroupError
}

func (nc *FakeNamespaceClient) CreateNamespace(ctx context.Context, azureCfg *azure.Config, groupName, namespace string, tags azure.Tags) (*eventhub.EHNamespace, error) {
	nc.Tags = tags
	return &eventhub.EHNamespace{
		Name: ptr.String(namespace),
	}, nc.PersistEventhubsNamespaceError
}

func (nc *FakeNamespaceClient) DeleteResourceGroup(ctx context.Context, tags azure.Tags) (resources.GroupsDeleteFuture, error) {
	//TODO(montaro) double check me
	nc.DeleteResourceGroupCalled = true
	return resources.GroupsDeleteFuture{}, nc.DeleteResourceGroupError
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

func NewFakeNamespaceClientResourceGroupDoesNotExist() *FakeNamespaceClient {
	return &FakeNamespaceClient{
		GetResourceGroupError: azure.NewResourceGroupDoesNotExist("ups .. resource group does not exist"),
	}
}

func NewFakeNamespaceClientResourceGroupConnectionError() *FakeNamespaceClient {
	return &FakeNamespaceClient{
		GetResourceGroupError: errors.New("ups .. can't connect to azure"),
	}
}

func NewFakeNamespaceClientResourceGroupDeleteError() *FakeNamespaceClient {
	return &FakeNamespaceClient{
		DeleteResourceGroupError: errors.New("error while trying to delete resource group"),
		GetResourceGroupReturnValue: resources.Group{
			Response:   autorest.Response{},
			Name:       ptr.String("montaro"),
			Properties: &resources.GroupProperties{ProvisioningState: ptr.String(azure.AzureFutureOperationSucceeded)},
		},
	}
}

func NewFakeNamespaceClientResourceGroupPropertiesError() *FakeNamespaceClient {
	return &FakeNamespaceClient{
		DeleteResourceGroupError: errors.New("error while trying to delete resource group"),
	}
}

func NewFakeNamespaceClientResourceGroupInDeletionMode() *FakeNamespaceClient {
	return &FakeNamespaceClient{
		GetResourceGroupReturnValue: resources.Group{
			Response:   autorest.Response{},
			Name:       ptr.String("montaro"),
			Properties: &resources.GroupProperties{ProvisioningState: ptr.String(azure.AzureFutureOperationDeleting)},
		},
	}
}

func NewFakeNamespaceClientResourceGroupExists() *FakeNamespaceClient {
	return &FakeNamespaceClient{
		GetResourceGroupReturnValue: resources.Group{
			Response:   autorest.Response{},
			Name:       ptr.String("montaro"),
			Properties: &resources.GroupProperties{ProvisioningState: ptr.String(azure.AzureFutureOperationSucceeded)},
		},
	}
}

// ensure the fake Client is implementing the interface
var _ azure.HyperscalerProvider = (*FakeHyperscalerProvider)(nil)

type FakeHyperscalerProvider struct {
	Client azure.AzureInterface
	Err    error
}

func (ac *FakeHyperscalerProvider) GetClient(config *azure.Config, logger logrus.FieldLogger) (azure.AzureInterface, error) {
	return ac.Client, ac.Err
}

func NewFakeHyperscalerProvider(client azure.AzureInterface) azure.HyperscalerProvider {
	return &FakeHyperscalerProvider{
		Client: client,
		Err:    nil,
	}
}

func NewFakeHyperscalerProviderError() azure.HyperscalerProvider {
	return &FakeHyperscalerProvider{
		Client: nil,
		Err:    fmt.Errorf("ups ... hyperscaler provider could not provide a hyperscaler Client"),
	}
}
