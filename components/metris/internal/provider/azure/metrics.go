package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/kyma-incubator/compass/components/metris/internal/metrics"
	"github.com/kyma-incubator/compass/components/metris/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	tagNameSubAccountID string = "SubAccountID"

	capMemoryGB = "MemoryGB"
	capvCPUs    = "vCPUs"

	diskSizeFactor float64 = 32

	intervalPT5M = 5 * time.Minute
)

// gatherMetrics - collect results from different Azure API and create edp events.
func (a *Provider) gatherMetrics(ctx context.Context, workerlogger *zap.SugaredLogger, shootname string) {
	defer utils.TrackTime("gatherMetrics", time.Now(), workerlogger)

	a.mu.RLock()
	client, ok := a.clients[shootname]
	a.mu.RUnlock()

	if !ok {
		workerlogger.Error("client config not found")
		return
	}

	var (
		subaccountid              = client.Account.SubAccountID
		resourceGroupName         = client.Account.TechnicalID
		eventHubResourceGroupName = ""
		event                     = &Event{
			Timestamp:      time.Now().Format(time.RFC3339),
			ResourceGroups: []string{resourceGroupName},
		}
	)

	metricTimer := prometheus.NewTimer(metrics.ReceivedSamplesDuration)
	defer metricTimer.ObserveDuration()

	workerlogger.Debug("getting metrics")

	rgFilter := fmt.Sprintf("tagname eq '%s' and tagvalue eq '%s'", tagNameSubAccountID, subaccountid)

	if ehResourceGroup, err := client.getResourceGroupList(ctx, rgFilter); err == nil && len(ehResourceGroup) > 0 {
		eventHubResourceGroupName = *ehResourceGroup[0].Name
		event.ResourceGroups = append(event.ResourceGroups, eventHubResourceGroupName)
	}

	event.Compute = a.getComputeMetrics(ctx, client, resourceGroupName, workerlogger)
	event.Networking = a.getNetworkMetrics(ctx, client, resourceGroupName, workerlogger)
	event.EventHub = a.getEventHubMetrics(ctx, client, eventHubResourceGroupName, workerlogger)

	events := make(Events)
	events[subaccountid] = make([]*Event, 0)
	events[subaccountid] = append(events[subaccountid], event)

	metrics.ReceivedSamples.Add(float64(len(events[subaccountid])))

	bufEvent, err := json.Marshal(events)
	if err != nil {
		workerlogger.Errorf("error parsing azure events to json: %s", err)
		return
	}

	a.eventsChannel <- &bufEvent
}

func (a *Provider) getComputeMetrics(ctx context.Context, client *Client, resourceGroupName string, logger *zap.SugaredLogger) Compute {
	defer utils.TrackTime("getComputeMetrics", time.Now(), logger)

	var (
		vms    []compute.VirtualMachine
		disks  []compute.Disk
		cpu    uint64
		ram    float64
		err    error
		result = Compute{
			VMTypes:          make([]VMType, 0),
			ProvisionedCpus:  0,
			ProvisionedRAMGB: 0,
			ProvisionedVolumes: ProvisionedVolume{
				SizeGBTotal:   0,
				SizeGBRounded: 0,
				Count:         0,
			},
		}
	)

	vms, err = client.getVirtualMachineList(ctx, resourceGroupName)
	if err != nil {
		logger.Warnf("could not get virtual machines information: %s", err)
	}

	vmt := make(map[string]uint32)
	for _, vm := range vms {
		vmt[string(vm.HardwareProfile.VMSize)]++

		capabilities, caperr := client.getVirtualMachineCapabilities(ctx, vm.HardwareProfile.VMSize)
		if caperr != nil {
			logger.Warnf("could not get vm capabilities: %s", err)
			continue
		}

		for _, v := range *capabilities {
			switch *v.Name {
			case capvCPUs:
				if cpu, err = strconv.ParseUint(*v.Value, 10, 32); err == nil {
					result.ProvisionedCpus += uint32(cpu)
				}
			case capMemoryGB:
				if ram, err = strconv.ParseFloat(*v.Value, 64); err == nil {
					result.ProvisionedRAMGB += ram
				}
			}
		}
	}

	for k, v := range vmt {
		result.VMTypes = append(result.VMTypes, VMType{Name: k, Count: v})
	}

	disks, err = client.getDiskList(ctx, resourceGroupName)
	if err != nil {
		logger.Warnf("could not get disk information: %s", err)
	} else {
		result.ProvisionedVolumes.Count = uint32(len(disks))

		for _, disk := range disks {
			result.ProvisionedVolumes.SizeGBTotal += uint32(*disk.DiskSizeGB)
			result.ProvisionedVolumes.SizeGBRounded += uint32(math.Ceil(float64(*disk.DiskSizeGB)/diskSizeFactor) * diskSizeFactor)
		}
	}

	return result
}

func (a *Provider) getNetworkMetrics(ctx context.Context, client *Client, resourceGroupName string, logger *zap.SugaredLogger) Networking {
	defer utils.TrackTime("getNetworkMetrics", time.Now(), logger)

	var (
		result = Networking{
			ProvisionedLoadBalancers: 0,
			ProvisionedIps:           0,
			ProvisionedVnets:         0,
		}
		err       error
		lbs       []network.LoadBalancer
		vnets     []network.VirtualNetwork
		publicIPs []network.PublicIPAddress
	)

	lbs, err = client.getLoadBalancerList(ctx, resourceGroupName)
	if err != nil {
		logger.Warnf("could not get loadbalancer infornation: %s", err)
	} else {
		result.ProvisionedLoadBalancers += uint32(len(lbs))
	}

	vnets, err = client.getVirtualNetworkList(ctx, resourceGroupName)
	if err != nil {
		logger.Warnf("could not get vnet infornation: %s", err)
	} else {
		result.ProvisionedVnets += uint32(len(vnets))
	}

	publicIPs, err = client.getPublicIPAddressList(ctx, resourceGroupName)
	if err != nil {
		logger.Warnf("could not get public ip infornation: %s", err)
	} else {
		result.ProvisionedIps += uint32(len(publicIPs))
	}

	return result
}

func (a *Provider) getEventHubMetrics(ctx context.Context, client *Client, resourceGroupName string, logger *zap.SugaredLogger) EventHub {
	defer utils.TrackTime("getEventHubMetrics", time.Now(), logger)

	var (
		result = EventHub{
			NumberNamespaces:     0,
			IncomingRequestsPT1M: 0,
			MaxIncomingBytesPT1M: 0,
			MaxOutgoingBytesPT1M: 0,
			IncomingRequestsPT5M: 0,
			MaxIncomingBytesPT5M: 0,
			MaxOutgoingBytesPT5M: 0,
		}
	)

	if resourceGroupName == "" {
		return result
	}

	ehns, eherr := client.getNamespaceList(ctx, resourceGroupName)
	if eherr != nil {
		logger.Warnf("eventhub namespace error: %s", eherr)
	}

	result.NumberNamespaces = uint32(len(ehns))

	interval := "PT1M"
	if a.pollinterval == intervalPT5M {
		interval = "PT5M"
	}

	for _, ns := range ehns {
		resourceURI := *ns.ID

		nsmetric, errs := client.getMetricValuesList(ctx, resourceURI, interval, []string{"IncomingBytes", "OutgoingBytes", "IncomingMessages"}, []string{string(insights.Maximum)})
		if len(errs) > 0 {
			for _, err := range errs {
				logger.Warnf("eventhub metric error: %s", err)
			}

			continue
		}

		if interval == "PT5M" {
			result.IncomingRequestsPT5M += *nsmetric["IncomingMessages"].Maximum
			result.MaxIncomingBytesPT5M += *nsmetric["IncomingBytes"].Maximum
			result.MaxOutgoingBytesPT5M += *nsmetric["OutgoingBytes"].Maximum
		} else {
			result.IncomingRequestsPT1M += *nsmetric["IncomingMessages"].Maximum
			result.MaxIncomingBytesPT1M += *nsmetric["IncomingBytes"].Maximum
			result.MaxOutgoingBytesPT1M += *nsmetric["OutgoingBytes"].Maximum
		}
	}

	return result
}
