package model

import "github.com/kyma-incubator/hydroform/types"

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

func (c GCPConfig) ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider) {
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
	}
	return cluster, provider
}

//func (db *deprovisioningBuilder) buildConfigForGCP(config model.GCPConfig) (*types.Cluster, *types.Provider, error) {

//}
