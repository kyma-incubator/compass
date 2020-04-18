package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

func (c *Client) getGroupClient() resources.GroupsClient {
	return resources.GroupsClient{BaseClient: *c.resourcesBaseClient}
}

func (c *Client) getVirtualMachinesClient() compute.VirtualMachinesClient {
	return compute.VirtualMachinesClient{BaseClient: *c.computeBaseClient}
}

func (c *Client) getDisksClient() compute.DisksClient {
	return compute.DisksClient{BaseClient: *c.computeBaseClient}
}

func (c *Client) getResourceSkusClient() compute.ResourceSkusClient {
	return compute.ResourceSkusClient{BaseClient: *c.computeBaseClient}
}

func (c *Client) getLoadBalancersClient() network.LoadBalancersClient {
	return network.LoadBalancersClient{BaseClient: *c.networkBaseClient}
}

func (c *Client) getVirtualNetworksClient() network.VirtualNetworksClient {
	return network.VirtualNetworksClient{BaseClient: *c.networkBaseClient}
}

func (c *Client) getPublicIPAddressesClient() network.PublicIPAddressesClient {
	return network.PublicIPAddressesClient{BaseClient: *c.networkBaseClient}
}

func (c *Client) getNamespacesClient() eventhub.NamespacesClient {
	return eventhub.NamespacesClient{BaseClient: *c.eventhubBaseClient}
}

func (c *Client) getMetricsClient() insights.MetricsClient {
	return insights.MetricsClient{BaseClient: *c.insightsBaseClient}
}
