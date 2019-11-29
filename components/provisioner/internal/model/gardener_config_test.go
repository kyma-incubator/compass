package model

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
)

func TestGardenerConfig_ToHydroformConfiguration(t *testing.T) {

	credentialsFile := "credentials"

	expectedCluster := &types.Cluster{
		Name:              "cluster",
		KubernetesVersion: "1.15",
		CPU:               0,
		DiskSizeGB:        30,
		NodeCount:         2,
		MachineType:       "machine",
		Location:          "eu",
		ClusterInfo:       nil,
	}

	getExpectedProvider := func(customConfig map[string]interface{}) *types.Provider {
		return &types.Provider{
			Type:                 types.Gardener,
			ProjectName:          "project",
			CredentialsFilePath:  credentialsFile,
			CustomConfigurations: customConfig,
		}
	}

	gcpGardenerInput := &gqlschema.GCPProviderConfigInput{Zone: "zone"}
	gcpGardenerProvider, err := NewGCPGardenerConfig(gcpGardenerInput)
	require.NoError(t, err)
	assert.Equal(t, `{"zone":"zone"}`, gcpGardenerProvider.RawJSON())
	assert.Equal(t, gcpGardenerInput, gcpGardenerProvider.input)

	azureGardenerInput := &gqlschema.AzureProviderConfigInput{VnetCidr: "10.10.11.11/255"}
	azureGardenerProvider, err := NewAzureGardenerConfig(azureGardenerInput)
	require.NoError(t, err)
	assert.Equal(t, `{"vnetCidr":"10.10.11.11/255"}`, azureGardenerProvider.RawJSON())
	assert.Equal(t, azureGardenerInput, azureGardenerProvider.input)

	awsGardenerInput := &gqlschema.AWSProviderConfigInput{
		Zone:         "zone",
		VpcCidr:      "10.10.11.11/255",
		PublicCidr:   "10.10.11.12/255",
		InternalCidr: "10.10.11.13/255",
	}
	awsGardenerProvider, err := NewAWSGardenerConfig(awsGardenerInput)
	require.NoError(t, err)
	expectedJSON := `{"zone":"zone","vpcCidr":"10.10.11.11/255","publicCidr":"10.10.11.12/255","internalCidr":"10.10.11.13/255"}`
	assert.Equal(t, expectedJSON, awsGardenerProvider.RawJSON())
	assert.Equal(t, awsGardenerInput, awsGardenerProvider.input)

	for _, testCase := range []struct {
		description                  string
		provider                     string
		providerConfig               GardenerProviderConfig
		expectedProviderCustomConfig map[string]interface{}
	}{
		{
			description:    "should convert to Hydroform config with GCP provider",
			provider:       "gcp",
			providerConfig: gcpGardenerProvider,
			expectedProviderCustomConfig: map[string]interface{}{
				"target_provider": "gcp",
				"target_seed":     "eu",
				"target_secret":   "gardener-secret",
				"disk_type":       "SSD",
				"workercidr":      "10.10.10.10/255",
				"autoscaler_min":  1,
				"autoscaler_max":  3,
				"max_surge":       30,
				"max_unavailable": 1,
				"zone":            "zone",
			},
		},
		{
			description:    "should convert to Hydroform config with Azure provider",
			provider:       "azure",
			providerConfig: azureGardenerProvider,
			expectedProviderCustomConfig: map[string]interface{}{
				"target_provider": "azure",
				"target_seed":     "eu",
				"target_secret":   "gardener-secret",
				"disk_type":       "SSD",
				"workercidr":      "10.10.10.10/255",
				"autoscaler_min":  1,
				"autoscaler_max":  3,
				"max_surge":       30,
				"max_unavailable": 1,
				"vnetcidr":        "10.10.11.11/255",
			},
		},
		{
			description:    "should convert to Hydroform config with AWS provider",
			provider:       "aws",
			providerConfig: awsGardenerProvider,
			expectedProviderCustomConfig: map[string]interface{}{
				"target_provider": "aws",
				"target_seed":     "eu",
				"target_secret":   "gardener-secret",
				"disk_type":       "SSD",
				"workercidr":      "10.10.10.10/255",
				"autoscaler_min":  1,
				"autoscaler_max":  3,
				"max_surge":       30,
				"max_unavailable": 1,
				"zone":            "zone",
				"vpccidr":         "10.10.11.11/255",
				"publicscidr":     "10.10.11.12/255",
				"internalscidr":   "10.10.11.13/255",
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerProviderConfig := fixGardenerConfig(testCase.provider, testCase.providerConfig)

			// when
			cluster, provider, err := gardenerProviderConfig.ToHydroformConfiguration(credentialsFile)

			// then
			require.NoError(t, err)
			expectedProvider := getExpectedProvider(testCase.expectedProviderCustomConfig)

			assert.Equal(t, expectedCluster, cluster)
			assert.Equal(t, expectedProvider, provider)
		})
	}
}

func Test_AsMap_Error(t *testing.T) {

	for _, testCase := range []struct {
		description            string
		gardenerProviderConfig GardenerProviderConfig
	}{
		{
			description:            "gcp gardener config",
			gardenerProviderConfig: &GCPGardenerConfig{ProviderSpecificConfig: ProviderSpecificConfig("invalid json")},
		},
		{
			description:            "azure gardener config",
			gardenerProviderConfig: &AzureGardenerConfig{ProviderSpecificConfig: ProviderSpecificConfig("invalid json")},
		},
		{
			description:            "azure gardener config",
			gardenerProviderConfig: &AWSGardenerConfig{ProviderSpecificConfig: ProviderSpecificConfig("invalid json")},
		},
	} {
		t.Run("should faild when invalid json for "+testCase.description, func(t *testing.T) {
			// when
			_, err := testCase.gardenerProviderConfig.AsMap()

			// then
			require.Error(t, err)
		})
	}

}

func fixGardenerConfig(provider string, providerCfg GardenerProviderConfig) GardenerConfig {
	return GardenerConfig{
		ID:                     "",
		ClusterID:              "",
		Name:                   "cluster",
		ProjectName:            "project",
		KubernetesVersion:      "1.15",
		NodeCount:              2,
		VolumeSizeGB:           30,
		DiskType:               "SSD",
		MachineType:            "machine",
		Provider:               provider,
		Seed:                   "eu",
		TargetSecret:           "gardener-secret",
		Region:                 "eu",
		WorkerCidr:             "10.10.10.10/255",
		AutoScalerMin:          1,
		AutoScalerMax:          3,
		MaxSurge:               30,
		MaxUnavailable:         1,
		GardenerProviderConfig: providerCfg,
	}
}
