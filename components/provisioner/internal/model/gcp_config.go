package model

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/hydroform/types"
)

type GCPConfig struct {
	ID                string
	ClusterID         string
	Name              string
	ProjectName       string
	KubernetesVersion string
	NumberOfNodes     int
	BootDiskSizeGB    int
	MachineType       string
	Region            string
	Zone              string
}

func (c GCPConfig) ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider, error) {
	cluster := &types.Cluster{
		KubernetesVersion: c.KubernetesVersion,
		Name:              c.Name,
		DiskSizeGB:        c.BootDiskSizeGB,
		NodeCount:         c.NumberOfNodes,
		Location:          c.Region,
		MachineType:       c.MachineType,
	}

	provider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         c.ProjectName,
		CredentialsFilePath: credentialsFileName,
		CustomConfigurations: map[string]interface{}{
			"zone": c.Zone,
		},
	}
	return cluster, provider, nil
}

func (c GCPConfig) ToShootTemplate(namespace string) (*gardener_types.Shoot, error) {
	panic("Method not supported for GCP Config")
}
