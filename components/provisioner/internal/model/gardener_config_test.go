package model

import (
	"fmt"
	"testing"

	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

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

	gcpGardenerProvider, err := NewGCPGardenerConfig(fixGCPGardenerInput())
	require.NoError(t, err)
	assert.Equal(t, `{"zone":"zone"}`, gcpGardenerProvider.RawJSON())
	assert.Equal(t, fixGCPGardenerInput(), gcpGardenerProvider.input)

	azureGardenerProvider, err := NewAzureGardenerConfig(fixAzureGardenerInput())
	require.NoError(t, err)
	assert.Equal(t, `{"vnetCidr":"10.10.11.11/255"}`, azureGardenerProvider.RawJSON())
	assert.Equal(t, fixAzureGardenerInput(), azureGardenerProvider.input)

	awsGardenerProvider, err := NewAWSGardenerConfig(fixAWSGardenerInput())
	require.NoError(t, err)
	expectedJSON := `{"zone":"zone","vpcCidr":"10.10.11.11/255","publicCidr":"10.10.11.12/255","internalCidr":"10.10.11.13/255"}`
	assert.Equal(t, expectedJSON, awsGardenerProvider.RawJSON())
	assert.Equal(t, fixAWSGardenerInput(), awsGardenerProvider.input)

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

func Test_NewGardenerConfigFromJSON(t *testing.T) {

	gcpConfigJSON := `{"zone":"zone"}`
	azureConfigJSON := `{"vnetCidr":"10.10.11.11/255"}`
	awsConfigJSON := `{"zone":"zone","vpcCidr":"10.10.11.11/255","publicCidr":"10.10.11.12/255","internalCidr":"10.10.11.13/255"}`

	for _, testCase := range []struct {
		description                    string
		jsonData                       string
		expectedConfig                 GardenerProviderConfig
		expectedProviderSpecificConfig gqlschema.ProviderSpecificConfig
	}{
		{
			description: "should create GCP Gardener config",
			jsonData:    gcpConfigJSON,
			expectedConfig: &GCPGardenerConfig{
				ProviderSpecificConfig: ProviderSpecificConfig(gcpConfigJSON),
				input:                  &gqlschema.GCPProviderConfigInput{Zone: "zone"},
			},
			expectedProviderSpecificConfig: gqlschema.GCPProviderConfig{Zone: util.StringPtr("zone")},
		},
		{
			description: "should create Azure Gardener config",
			jsonData:    azureConfigJSON,
			expectedConfig: &AzureGardenerConfig{
				ProviderSpecificConfig: ProviderSpecificConfig(azureConfigJSON),
				input:                  &gqlschema.AzureProviderConfigInput{VnetCidr: "10.10.11.11/255"},
			},
			expectedProviderSpecificConfig: gqlschema.AzureProviderConfig{VnetCidr: util.StringPtr("10.10.11.11/255")},
		},
		{
			description: "should create AWS Gardener config",
			jsonData:    awsConfigJSON,
			expectedConfig: &AWSGardenerConfig{
				ProviderSpecificConfig: ProviderSpecificConfig(awsConfigJSON),
				input: &gqlschema.AWSProviderConfigInput{
					Zone:         "zone",
					VpcCidr:      "10.10.11.11/255",
					PublicCidr:   "10.10.11.12/255",
					InternalCidr: "10.10.11.13/255",
				},
			},
			expectedProviderSpecificConfig: gqlschema.AWSProviderConfig{
				Zone:         util.StringPtr("zone"),
				VpcCidr:      util.StringPtr("10.10.11.11/255"),
				PublicCidr:   util.StringPtr("10.10.11.12/255"),
				InternalCidr: util.StringPtr("10.10.11.13/255"),
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// when
			gardenerProviderConfig, err := NewGardenerProviderConfigFromJSON(testCase.jsonData)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedConfig, gardenerProviderConfig)

			// when
			providerSpecificConfig := gardenerProviderConfig.AsProviderSpecificConfig()

			// then
			assert.Equal(t, testCase.expectedProviderSpecificConfig, providerSpecificConfig)
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

func TestGardenerConfig_ToShootTemplate(t *testing.T) {

	gcpGardenerProvider, err := NewGCPGardenerConfig(fixGCPGardenerInput())
	require.NoError(t, err)

	azureGardenerProvider, err := NewAzureGardenerConfig(fixAzureGardenerInput())
	require.NoError(t, err)

	awsGardenerProvider, err := NewAWSGardenerConfig(fixAWSGardenerInput())
	require.NoError(t, err)

	for _, testCase := range []struct {
		description           string
		provider              string
		providerConfig        GardenerProviderConfig
		expectedShootTemplate *gardener_types.Shoot
	}{
		{
			description:    "should convert to Shoot template with GCP provider",
			provider:       "gcp",
			providerConfig: gcpGardenerProvider,
			expectedShootTemplate: &gardener_types.Shoot{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cluster",
					Namespace: "gardener-namespace",
				},
				Spec: gardener_types.ShootSpec{
					CloudProfileName: "gcp",
					Networking: gardener_types.Networking{
						Type:  "calico",
						Nodes: "10.250.0.0/19",
					},
					SeedName:          util.StringPtr("eu"),
					SecretBindingName: "gardener-secret",
					Region:            "eu",
					Provider: gardener_types.Provider{
						Type: "gcp",
						ControlPlaneConfig: &gardener_types.ProviderConfig{
							RawExtension: apimachineryRuntime.RawExtension{
								Raw: []byte(`{"kind":"ControlPlaneConfig","apiVersion":"gcp.provider.extensions.gardener.cloud/v1alpha1","zone":"zone"}`)},
						},
						InfrastructureConfig: &gardener_types.ProviderConfig{
							RawExtension: apimachineryRuntime.RawExtension{
								Raw: []byte(`{"kind":"InfrastructureConfig","apiVersion":"gcp.provider.extensions.gardener.cloud/v1alpha1","networks":{"worker":"10.10.10.10/255","workers":"10.10.10.10/255"}}`),
							},
						},
						Workers: []gardener_types.Worker{
							fixWorker(0, []string{"zone"}),
							fixWorker(1, []string{"zone"}),
						},
					},
					Kubernetes: gardener_types.Kubernetes{
						AllowPrivilegedContainers: util.BoolPtr(true),
						Version:                   "1.15",
						KubeAPIServer: &gardener_types.KubeAPIServerConfig{
							EnableBasicAuthentication: util.BoolPtr(false),
						},
					},
					Maintenance: &gardener_types.Maintenance{},
				},
			},
		},
		{
			description:    "should convert to Shoot template with Azure provider",
			provider:       "az",
			providerConfig: azureGardenerProvider,
			expectedShootTemplate: &gardener_types.Shoot{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cluster",
					Namespace: "gardener-namespace",
				},
				Spec: gardener_types.ShootSpec{
					CloudProfileName: "az",
					Networking: gardener_types.Networking{
						Type:  "calico",
						Nodes: "10.250.0.0/19",
					},
					SeedName:          util.StringPtr("eu"),
					SecretBindingName: "gardener-secret",
					Region:            "eu",
					Provider: gardener_types.Provider{
						Type: "azure",
						ControlPlaneConfig: &gardener_types.ProviderConfig{
							RawExtension: apimachineryRuntime.RawExtension{
								Raw: []byte(`{"kind":"ControlPlaneConfig","apiVersion":"azure.provider.extensions.gardener.cloud/v1alpha1"}`)},
						},
						InfrastructureConfig: &gardener_types.ProviderConfig{
							RawExtension: apimachineryRuntime.RawExtension{
								Raw: []byte(`{"kind":"InfrastructureConfig","apiVersion":"azure.provider.extensions.gardener.cloud/v1alpha1","networks":{"vnet":{"cidr":"10.10.11.11/255"},"workers":"10.10.10.10/255"},"zoned":false}`),
							},
						},
						Workers: []gardener_types.Worker{
							fixWorker(0, nil),
							fixWorker(1, nil),
						},
					},
					Kubernetes: gardener_types.Kubernetes{
						AllowPrivilegedContainers: util.BoolPtr(true),
						Version:                   "1.15",
						KubeAPIServer: &gardener_types.KubeAPIServerConfig{
							EnableBasicAuthentication: util.BoolPtr(false),
						},
					},
					Maintenance: &gardener_types.Maintenance{},
				},
			},
		},
		{
			description:    "should convert to Shoot template with AWS provider",
			provider:       "aws",
			providerConfig: awsGardenerProvider,
			expectedShootTemplate: &gardener_types.Shoot{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cluster",
					Namespace: "gardener-namespace",
				},
				Spec: gardener_types.ShootSpec{
					CloudProfileName: "aws",
					Networking: gardener_types.Networking{
						Type:  "calico",
						Nodes: "10.250.0.0/19",
					},
					SeedName:          util.StringPtr("eu"),
					SecretBindingName: "gardener-secret",
					Region:            "eu",
					Provider: gardener_types.Provider{
						Type: "aws",
						ControlPlaneConfig: &gardener_types.ProviderConfig{
							RawExtension: apimachineryRuntime.RawExtension{
								Raw: []byte(`{"kind":"ControlPlaneConfig","apiVersion":"aws.provider.extensions.gardener.cloud/v1alpha1"}`)},
						},
						InfrastructureConfig: &gardener_types.ProviderConfig{
							RawExtension: apimachineryRuntime.RawExtension{
								Raw: []byte(`{"kind":"InfrastructureConfig","apiVersion":"aws.provider.extensions.gardener.cloud/v1alpha1","networks":{"vpc":{"cidr":"10.10.11.11/255"},"zones":[{"name":"zone","internal":"10.10.11.13/255","public":"10.10.11.12/255","workers":"10.10.10.10/255"}]}}`),
							},
						},
						Workers: []gardener_types.Worker{
							fixWorker(0, []string{"zone"}),
							fixWorker(1, []string{"zone"}),
						},
					},
					Kubernetes: gardener_types.Kubernetes{
						AllowPrivilegedContainers: util.BoolPtr(true),
						Version:                   "1.15",
						KubeAPIServer: &gardener_types.KubeAPIServerConfig{
							EnableBasicAuthentication: util.BoolPtr(false),
						},
					},
					Maintenance: &gardener_types.Maintenance{},
				},
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			gardenerProviderConfig := fixGardenerConfig(testCase.provider, testCase.providerConfig)

			// when
			template, err := gardenerProviderConfig.ToShootTemplate("gardener-namespace")

			// then
			fmt.Println(string(template.Spec.Provider.InfrastructureConfig.Raw))
			fmt.Println(string(template.Spec.Provider.ControlPlaneConfig.Raw))

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedShootTemplate, template)
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

func fixAWSGardenerInput() *gqlschema.AWSProviderConfigInput {
	return &gqlschema.AWSProviderConfigInput{
		Zone:         "zone",
		VpcCidr:      "10.10.11.11/255",
		PublicCidr:   "10.10.11.12/255",
		InternalCidr: "10.10.11.13/255",
	}
}

func fixGCPGardenerInput() *gqlschema.GCPProviderConfigInput {
	return &gqlschema.GCPProviderConfigInput{Zone: "zone"}
}

func fixAzureGardenerInput() *gqlschema.AzureProviderConfigInput {
	return &gqlschema.AzureProviderConfigInput{VnetCidr: "10.10.11.11/255"}
}

func fixWorker(index int, zones []string) gardener_types.Worker {
	return gardener_types.Worker{
		Name:           fmt.Sprintf("cpu-worker-%d", index),
		MaxSurge:       util.IntOrStrPtr(intstr.FromInt(30)),
		MaxUnavailable: util.IntOrStrPtr(intstr.FromInt(1)),
		Machine: gardener_types.Machine{
			Type: "machine",
		},
		Volume: &gardener_types.Volume{
			Type: util.StringPtr("SSD"),
			Size: "30Gi",
		},
		Maximum: 3,
		Minimum: 1,
		Zones:   zones,
	}
}
