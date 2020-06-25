package model

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
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

func (c GCPConfig) ToShootTemplate(namespace string, accountId string, subAccountId string) (*gardener_types.Shoot, error) {
	panic("Method not supported for GCP Config")
}
