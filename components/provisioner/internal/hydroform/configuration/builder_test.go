package configuration

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	v1 "k8s.io/api/core/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	secretName = "gardener"
	namespace  = "compass-system"
)

func TestProvisioningBuilder(t *testing.T) {
	t.Run("Should return valid GCP configuration", func(t *testing.T) {
		//given
		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: &gqlschema.ClusterConfigInput{GcpConfig: &gqlschema.GCPConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			NumberOfNodes:     3,
			BootDiskSizeGb:    256,
			MachineType:       "n1-standard-1",
			Region:            "region",
			KubernetesVersion: "version",
		}},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: secretName,
			},
		}

		expectedProvider := &types.Provider{
			Type:                 types.GCP,
			ProjectName:          "Project",
			CustomConfigurations: nil,
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewProvisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})

	t.Run("Should return valid GCP Gardener configuration", func(t *testing.T) {
		//given
		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: &gqlschema.ClusterConfigInput{GardenerConfig: &gqlschema.GardenerConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			MachineType:       "n1-standard-1",
			Region:            "region",
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      256,
			DiskType:          "ssd",
			Provider:          "GCP",
			Seed:              "gcp-eu1",
			TargetSecret:      "secret",
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zone: "zone",
				},
			},
		}},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: secretName,
			},
		}

		expectedProvider := &types.Provider{
			Type:        types.Gardener,
			ProjectName: "Project",
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"workercidr":      "cidr",
				"disk_type":       "ssd",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "GCP",
				"target_seed":     "gcp-eu1",
				"target_secret":   "secret",
				"zone":            "zone"},
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewProvisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})

	t.Run("Should return valid Azure Gardener configuration", func(t *testing.T) {
		//given
		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: &gqlschema.ClusterConfigInput{GardenerConfig: &gqlschema.GardenerConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			MachineType:       "n1-standard-1",
			Region:            "region",
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      256,
			DiskType:          "standard",
			Provider:          "Azure",
			Seed:              "az-eu1",
			TargetSecret:      "secret",
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AzureConfig: &gqlschema.AzureProviderConfigInput{
					VnetCidr: "cidr",
				},
			},
		}},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: secretName,
			},
		}

		expectedProvider := &types.Provider{
			Type:        types.Gardener,
			ProjectName: "Project",
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"workercidr":      "cidr",
				"disk_type":       "standard",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "Azure",
				"target_seed":     "az-eu1",
				"target_secret":   "secret",
				"vnetcidr":        "cidr"},
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewProvisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})

	t.Run("Should return valid AWS Gardener configuration", func(t *testing.T) {
		//given
		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: &gqlschema.ClusterConfigInput{GardenerConfig: &gqlschema.GardenerConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			MachineType:       "n1-standard-1",
			Region:            "region",
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      256,
			DiskType:          "standard",
			Provider:          "aws",
			Seed:              "aws-eu1",
			TargetSecret:      "secret",
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AwsConfig: &gqlschema.AWSProviderConfigInput{
					Zone:         "zone",
					PublicCidr:   "cidr",
					InternalCidr: "cidr",
					VpcCidr:      "cidr",
				},
			},
		}},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: secretName,
			},
		}

		expectedProvider := &types.Provider{
			Type:        types.Gardener,
			ProjectName: "Project",
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"workercidr":      "cidr",
				"disk_type":       "standard",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "aws",
				"target_seed":     "aws-eu1",
				"target_secret":   "secret",
				"zone":            "zone",
				"internalscidr":   "cidr",
				"vpccidr":         "cidr",
				"publicscidr":     "cidr",
			}}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewProvisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})
}

func TestDeprovisioningBuilder(t *testing.T) {
	t.Run("Should return valid GCP configuration", func(t *testing.T) {
		//given
		config := model.RuntimeConfig{ClusterConfig: model.GCPConfig{
			ID:                "id",
			Name:              "Something",
			ProjectName:       "Project",
			NumberOfNodes:     3,
			BootDiskSizeGB:    256,
			MachineType:       "n1-standard-1",
			Region:            "region",
			KubernetesVersion: "version",
			ClusterID:         "runtimeID",
		},
			CredentialsSecretName: secretName,
		}

		expectedProvider := &types.Provider{
			Type:                 types.GCP,
			ProjectName:          "Project",
			CustomConfigurations: nil,
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewDeprovisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})

	t.Run("Should return valid GCP Gardener configuration", func(t *testing.T) {
		//given
		config := model.RuntimeConfig{ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "Something",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           256,
			DiskType:               "ssd",
			Provider:               "GCP",
			Seed:                   "gcp-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			ProviderSpecificConfig: "{\"zone\":\"zone\"}",
		},
			CredentialsSecretName: secretName,
		}

		expectedProvider := &types.Provider{
			Type:        types.Gardener,
			ProjectName: "Project",
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"workercidr":      "cidr",
				"disk_type":       "ssd",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "GCP",
				"target_seed":     "gcp-eu1",
				"target_secret":   "secret",
				"zone":            "zone"},
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewDeprovisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})

	t.Run("Should return valid Azure Gardener configuration", func(t *testing.T) {
		//given
		config := model.RuntimeConfig{ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "Something",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           256,
			DiskType:               "standard",
			Provider:               "azure",
			Seed:                   "az-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			ProviderSpecificConfig: "{\"vnetCidr\":\"cidr\"}",
		},
			CredentialsSecretName: secretName,
		}

		expectedProvider := &types.Provider{
			Type:        types.Gardener,
			ProjectName: "Project",
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"workercidr":      "cidr",
				"disk_type":       "standard",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "azure",
				"target_seed":     "az-eu1",
				"target_secret":   "secret",
				"vnetcidr":        "cidr",
			},
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewDeprovisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})

	t.Run("Should return valid AWS Gardener configuration", func(t *testing.T) {
		//given
		config := model.RuntimeConfig{ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "Something",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           256,
			DiskType:               "standard",
			Provider:               "aws",
			Seed:                   "aws-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			ProviderSpecificConfig: "{\"zone\":\"zone\", \"internalCidr\":\"cidr\", \"vpcCidr\":\"cidr\", \"publicCidr\":\"cidr\"}",
		},
			CredentialsSecretName: secretName,
		}

		expectedProvider := &types.Provider{
			Type:        types.Gardener,
			ProjectName: "Project",
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"workercidr":      "cidr",
				"disk_type":       "standard",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "aws",
				"target_seed":     "aws-eu1",
				"target_secret":   "secret",
				"zone":            "zone",
				"internalscidr":   "cidr",
				"vpccidr":         "cidr",
				"publicscidr":     "cidr",
			},
		}

		expectedCluster := &types.Cluster{
			Name:              "Something",
			NodeCount:         3,
			DiskSizeGB:        256,
			MachineType:       "n1-standard-1",
			Location:          "region",
			KubernetesVersion: "version",
		}

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		factory := NewConfigBuilderFactory(secrets)

		builder := factory.NewDeprovisioningBuilder(config)

		//when
		cluster, provider, err := builder.Create()

		//then
		require.NoError(t, err)
		providersEqual(t, expectedProvider, provider)
		assert.Equal(t, expectedCluster, cluster)

		//cleanup
		builder.CleanUp()
	})
}

func providersEqual(t *testing.T, expectedProvider, actualProvider *types.Provider) {
	assert.Equal(t, expectedProvider.Type, actualProvider.Type)
	assert.Equal(t, expectedProvider.CustomConfigurations, actualProvider.CustomConfigurations)
	assert.Equal(t, expectedProvider.ProjectName, actualProvider.ProjectName)
	assert.NotEmpty(t, actualProvider.CredentialsFilePath)
}

func createFakeCredentialsSecret(t *testing.T, secrets core.SecretInterface) {
	secret := &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			credentialsKey: []byte("YmFzZTY0IGNyZWRlbnRpYWxz"),
		},
	}

	_, err := secrets.Create(secret)

	require.NoError(t, err)
}

func deleteSecret(t *testing.T, secrets core.SecretInterface) {
	err := secrets.Delete(secretName, &meta.DeleteOptions{})
	require.NoError(t, err)
}
