package provider

import (
	"fmt"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
)

const (
	DefaultGCPRegion = "europe-west4"
)

type GcpInput struct{}

func (p *GcpInput) Defaults() *gqlschema.ClusterConfigInput {
	return &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			DiskType:       "pd-standard",
			VolumeSizeGb:   30,
			MachineType:    "n1-standard-4",
			Region:         DefaultGCPRegion,
			Provider:       "gcp",
			WorkerCidr:     "10.250.0.0/19",
			AutoScalerMin:  3,
			AutoScalerMax:  4,
			MaxSurge:       4,
			MaxUnavailable: 1,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zones: ZonesForGCPRegion(DefaultGCPRegion),
				},
			},
		},
	}
}

func (p *GcpInput) ApplyParameters(input *gqlschema.ClusterConfigInput, params internal.ProvisioningParametersDTO) {
	if params.Region != nil && params.Zones == nil {
		updateSlice(&input.GardenerConfig.ProviderSpecificConfig.GcpConfig.Zones, ZonesForGCPRegion(*params.Region))
	}

	updateSlice(&input.GardenerConfig.ProviderSpecificConfig.GcpConfig.Zones, params.Zones)
}

func ZonesForGCPRegion(region string) []string {
	var zones []string

	for _, name := range []string{"a", "b", "c"} {
		zones = append(zones, fmt.Sprintf("%s-%s", region, name))
	}

	return zones
}
