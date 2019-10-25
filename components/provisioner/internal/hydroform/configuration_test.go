package hydroform

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguration(t *testing.T) {
	t.Run("Should return correct gcp configuration", func(t *testing.T) {
		//given
		config := model.RuntimeConfig{ClusterConfig: model.GCPConfig{
			ID:                "id",
			Name:              "Something",
			ProjectName:       "Project",
			NumberOfNodes:     3,
			BootDiskSize:      "256",
			MachineType:       "n1-standard-1",
			Region:            "region",
			KubernetesVersion: "version",
			ClusterID:         "runtimeID",
		}}

		credentials := "credentials.yaml"

		expectedProvider := &types.Provider{
			Type:                 types.GCP,
			ProjectName:          "Project",
			CredentialsFilePath:  credentials,
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

		//when
		cluster, provider, err := prepareConfig(config, credentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedCluster, cluster)
		assert.Equal(t, expectedProvider, provider)
	})

	t.Run("Should return correct gardener configuration", func(t *testing.T) {
		//given
		config := model.RuntimeConfig{ClusterConfig: model.GardenerConfig{
			ID:                "id",
			Name:              "Something",
			ProjectName:       "Project",
			MachineType:       "n1-standard-1",
			Region:            "region",
			Zone:              "zone",
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSize:        "256",
			DiskType:          "ssd",
			TargetProvider:    "GCP",
			TargetSecret:      "secret",
			Cidr:              "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ClusterID:         "runtimeID",
		}}

		credentials := "credentials.yaml"

		expectedProvider := &types.Provider{
			Type:                types.Gardener,
			ProjectName:         "Project",
			CredentialsFilePath: credentials,
			CustomConfigurations: map[string]interface{}{
				"autoscaler_max":  5,
				"autoscaler_min":  1,
				"cidr":            "cidr",
				"disk_type":       "ssd",
				"max_surge":       1,
				"max_unavailable": 2,
				"target_provider": "GCP",
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

		//when
		cluster, provider, err := prepareConfig(config, credentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedCluster, cluster)
		assert.Equal(t, expectedProvider, provider)
	})
}
