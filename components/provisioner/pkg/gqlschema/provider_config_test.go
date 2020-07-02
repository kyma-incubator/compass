package gqlschema

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyma-project/control-plane/components/provisioner/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGardenerConfig_UnmarshalJSON(t *testing.T) {

	azureProviderCfgNoZones := &AzureProviderConfig{VnetCidr: util.StringPtr("10.10.11.11/25")}
	azureProviderCfg := &AzureProviderConfig{VnetCidr: util.StringPtr("10.10.11.11/25"), Zones: []string{"az-zone-1", "az-zone-2"}}
	gcpProviderCfg := &GCPProviderConfig{Zones: []string{"gcp-zone-1", "gcp-zone-2"}}
	awsProviderCfg := &AWSProviderConfig{
		Zone:         util.StringPtr("aws zone"),
		VpcCidr:      util.StringPtr("10.10.10.11/25"),
		PublicCidr:   util.StringPtr("10.10.10.12/25"),
		InternalCidr: util.StringPtr("10.10.10.13/25"),
	}

	for _, testCase := range []struct {
		description    string
		gardenerConfig GardenerConfig
	}{
		{
			description:    "gardener cluster with Azure with no zones passed",
			gardenerConfig: newGardenerClusterCfg(fixGardenerConfig("azure"), azureProviderCfgNoZones),
		},
		{
			description:    "gardener cluster with Azure",
			gardenerConfig: newGardenerClusterCfg(fixGardenerConfig("azure"), azureProviderCfg),
		},
		{
			description:    "gardener cluster with GCP",
			gardenerConfig: newGardenerClusterCfg(fixGardenerConfig("gcp"), gcpProviderCfg),
		},
		{
			description:    "gardener cluster with AWS",
			gardenerConfig: newGardenerClusterCfg(fixGardenerConfig("aws"), awsProviderCfg),
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			marshalled, err := json.Marshal(testCase.gardenerConfig)
			require.NoError(t, err)

			var unmarshalledConfig GardenerConfig

			// when
			err = json.NewDecoder(bytes.NewBuffer(marshalled)).Decode(&unmarshalledConfig)
			require.NoError(t, err)

			// then
			assert.Equal(t, testCase.gardenerConfig, unmarshalledConfig)
		})
	}

}

func newGardenerClusterCfg(gardenerCfg GardenerConfig, providerCfg ProviderSpecificConfig) GardenerConfig {
	gardenerCfg.ProviderSpecificConfig = providerCfg

	return gardenerCfg
}

func fixGardenerConfig(providerName string) GardenerConfig {
	return GardenerConfig{
		Name:              util.StringPtr("name"),
		KubernetesVersion: util.StringPtr("1.16"),
		VolumeSizeGb:      util.IntPtr(50),
		MachineType:       util.StringPtr("machine"),
		Region:            util.StringPtr("region"),
		Provider:          util.StringPtr(providerName),
		Seed:              util.StringPtr("seed"),
		TargetSecret:      util.StringPtr("secret"),
		DiskType:          util.StringPtr("disk"),
		WorkerCidr:        util.StringPtr("10.10.10.10/25"),
		AutoScalerMin:     util.IntPtr(1),
		AutoScalerMax:     util.IntPtr(4),
		MaxSurge:          util.IntPtr(25),
		MaxUnavailable:    util.IntPtr(2),
	}
}
