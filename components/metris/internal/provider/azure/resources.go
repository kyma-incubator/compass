package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

func (c *Client) getResourceGroupList(ctx context.Context, filter string) ([]resources.Group, error) {
	var results []resources.Group

	rgClient := c.getGroupClient()

	for rgList, err := rgClient.ListComplete(ctx, filter, nil); rgList.NotDone(); err = rgList.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}

		results = append(results, rgList.Value())
	}

	return results, nil
}

func (c *Client) getResourceGroup(ctx context.Context, rgname string) (resources.Group, error) {
	rgClient := c.getGroupClient()

	result, err := rgClient.Get(ctx, rgname)
	if err != nil {
		return resources.Group{}, err
	}

	return result, nil
}

func (c *Client) initVirtualMachineCapabilities(ctx context.Context) error {
	ResourceSkusList.mu.Lock()
	defer ResourceSkusList.mu.Unlock()

	if len(ResourceSkusList.skus) == 0 {
		c.logger.Debug("resourceSkus cache is empty, initializing it")

		skuClient := c.getResourceSkusClient()

		filter := ""

		if len(c.Location) > 0 {
			filter = fmt.Sprintf("location eq '%s'", c.Location)
		}

		for skuList, err := skuClient.ListComplete(ctx, filter); skuList.NotDone(); err = skuList.NextWithContext(ctx) {
			if err != nil {
				return err
			}

			item := skuList.Value()
			if *item.ResourceType == "virtualMachines" {
				if _, ok := ResourceSkusList.skus[*item.Name]; !ok {
					ResourceSkusList.skus[*item.Name] = &item
				}
			}
		}
	}

	return nil
}

func (c *Client) getVirtualMachineCapabilities(ctx context.Context, sizeType compute.VirtualMachineSizeTypes) (*[]compute.ResourceSkuCapabilities, error) {
	var result *[]compute.ResourceSkuCapabilities

	skuName := string(sizeType)

	if len(ResourceSkusList.skus) == 0 {
		if err := c.initVirtualMachineCapabilities(ctx); err != nil {
			return result, fmt.Errorf("couldn't initialize capabilities %s", err)
		}
	}

	ResourceSkusList.mu.Lock()
	defer ResourceSkusList.mu.Unlock()

	if rs, ok := ResourceSkusList.skus[skuName]; ok {
		result = rs.Capabilities
	}

	if len(*result) == 0 {
		return result, fmt.Errorf("couldn't find capabilities for machine type '%s'", skuName)
	}

	return result, nil
}

func (c *Client) getVirtualMachineList(ctx context.Context, resourceGroupName string) ([]compute.VirtualMachine, error) {
	var result []compute.VirtualMachine

	vmClient := c.getVirtualMachinesClient()

	for vmList, err := vmClient.ListComplete(ctx, resourceGroupName); vmList.NotDone(); err = vmList.NextWithContext(ctx) {
		if err != nil {
			return result, err
		}

		result = append(result, vmList.Value())
	}

	return result, nil
}

func (c *Client) getDiskList(ctx context.Context, resourceGroupName string) ([]compute.Disk, error) {
	var result []compute.Disk

	diskClient := c.getDisksClient()

	for diskList, err := diskClient.ListByResourceGroup(ctx, resourceGroupName); diskList.NotDone(); err = diskList.NextWithContext(ctx) {
		if err != nil {
			return result, err
		}

		for _, disk := range diskList.Values() {
			if len(disk.DiskProperties.OsType) == 0 {
				result = append(result, disk)
			}
		}
	}

	return result, nil
}

func (c *Client) getLoadBalancerList(ctx context.Context, resourceGroupName string) ([]network.LoadBalancer, error) {
	var result []network.LoadBalancer

	loadBalancersClient := c.getLoadBalancersClient()

	for lbList, err := loadBalancersClient.ListComplete(ctx, resourceGroupName); lbList.NotDone(); err = lbList.NextWithContext(ctx) {
		if err != nil {
			return result, err
		}

		result = append(result, lbList.Value())
	}

	return result, nil
}

func (c *Client) getVirtualNetworkList(ctx context.Context, resourceGroupName string) ([]network.VirtualNetwork, error) {
	var result []network.VirtualNetwork

	virtualNetworksClient := c.getVirtualNetworksClient()

	for vnetList, err := virtualNetworksClient.ListComplete(ctx, resourceGroupName); vnetList.NotDone(); err = vnetList.NextWithContext(ctx) {
		if err != nil {
			return result, err
		}

		result = append(result, vnetList.Value())
	}

	return result, nil
}

func (c *Client) getPublicIPAddressList(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, error) {
	var result []network.PublicIPAddress

	publicIPAddressesClient := c.getPublicIPAddressesClient()

	for ipList, err := publicIPAddressesClient.ListComplete(ctx, resourceGroupName); ipList.NotDone(); err = ipList.NextWithContext(ctx) {
		if err != nil {
			return result, err
		}

		result = append(result, ipList.Value())
	}

	return result, nil
}

func (c *Client) getNamespaceList(ctx context.Context, resourcegroup string) ([]eventhub.EHNamespace, error) {
	var results []eventhub.EHNamespace

	nsClient := c.getNamespacesClient()

	for nsList, err := nsClient.ListByResourceGroupComplete(ctx, resourcegroup); nsList.NotDone(); err = nsList.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}

		results = append(results, nsList.Value())
	}

	return results, nil
}

// getMetricValuesList returns the specified metric data points for the specified resource ID spanning the last 5 minutes.
func (c *Client) getMetricValuesList(ctx context.Context, resourceURI, interval string, metricnames, aggregations []string) (map[string]insights.MetricValue, []error) {
	var (
		results = make(map[string]insights.MetricValue)
		errors  []error
	)

	// interval possible values: PT1M, PT5M, PT15M, PT30M, PT1H

	endTime := time.Now().UTC()
	startTime := endTime.Add(time.Duration(-5) * time.Minute)
	timespan := fmt.Sprintf("%s/%s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	if len(aggregations) == 0 {
		aggregations = []string{
			string(insights.Average),
			string(insights.Count),
			string(insights.Maximum),
			string(insights.Minimum),
			string(insights.Total),
		}
	}

	metricsClient := c.getMetricsClient()

	metricsList, err := metricsClient.List(ctx, resourceURI, timespan, &interval, strings.Join(metricnames, ","), strings.Join(aggregations, ","), nil, "", "", insights.Data, "")
	if err != nil {
		return nil, append(errors, err)
	}

	for _, metric := range *metricsList.Value {
		metricName := *metric.Name.Value
		ts := *metric.Timeseries

		if len(ts) == 0 {
			errors = append(errors, fmt.Errorf("metric %s not found at target %s", metricName, *metric.ID))
			continue
		}

		tsdata := *ts[0].Data
		if len(tsdata) == 0 {
			errors = append(errors, fmt.Errorf("no metric data returned for metric %s at target %s", metricName, *metric.ID))
			continue
		}

		metricValue := tsdata[len(tsdata)-1]
		results[metricName] = metricValue
	}

	return results, errors
}
