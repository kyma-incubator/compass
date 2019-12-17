package model

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/hydroform/types"
)

func TestGCPConfig_ToHydroformConfiguration(t *testing.T) {

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

	expectedProvider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         "project",
		CredentialsFilePath: credentialsFile,
		CustomConfigurations: map[string]interface{}{
			"zone": "west",
		},
	}

	gcpConfig := GCPConfig{
		ID:                "",
		ClusterID:         "",
		Name:              "cluster",
		ProjectName:       "project",
		KubernetesVersion: "1.15",
		NumberOfNodes:     2,
		BootDiskSizeGB:    30,
		MachineType:       "machine",
		Region:            "eu",
		Zone:              "west",
	}

	cluster, provider, err := gcpConfig.ToHydroformConfiguration(credentialsFile)

	require.NoError(t, err)
	assert.Equal(t, expectedCluster, cluster)
	assert.Equal(t, expectedProvider, provider)
}
