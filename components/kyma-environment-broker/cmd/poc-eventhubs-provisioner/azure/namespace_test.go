package azure

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/go-autorest/autorest"
)

func TestGetResourceGroup(t *testing.T) {

	resourceGroup := "my-resourcegroup"
	eventHubNamespaceID := fmt.Sprintf("/subscriptions/35d42578-34d1-486d-a689-012a8d514c19/resourceGroups/%s/providers/Microsoft.EventHub/namespaces/nachtmaar-test-2", resourceGroup)
	nameSpaceFixture := eventhub.EHNamespace{
		Response: autorest.Response{},
		ID:       &eventHubNamespaceID,
	}
	parsedResourceGroup := GetResourceGroup(nameSpaceFixture)
	if parsedResourceGroup != resourceGroup {
		t.Errorf("ResourceGroup should be %q, but is: %q", resourceGroup, parsedResourceGroup)
	}
}
