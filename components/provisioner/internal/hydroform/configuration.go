package hydroform

import (
	"errors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"strconv"
)

func prepareConfig(input model.RuntimeConfig, credentialsFile string) (*types.Cluster, *types.Provider, error) {
	gardenerConfig, ok := input.GardenerConfig()
	if ok {
		return buildConfigForGardener(gardenerConfig, credentialsFile)
	}

	gcpConfig, ok := input.GCPConfig()
	if ok {
		return buildConfigForGCP(gcpConfig, credentialsFile)
	}

	return nil, nil, errors.New("configuration does not match any provider profiles")
}

func buildConfigForGCP(config model.GCPConfig, credentialsFile string) (*types.Cluster, *types.Provider, error) {
	diskSize, err := strconv.Atoi(config.BootDiskSize)

	if err != nil {
		return &types.Cluster{}, &types.Provider{}, err
	}

	cluster := &types.Cluster{
		KubernetesVersion: config.KubernetesVersion,
		Name:              config.Name,
		DiskSizeGB:        diskSize,
		NodeCount:         config.NumberOfNodes,
		Location:          config.Region,
		MachineType:       config.MachineType,
	}

	provider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         config.ProjectName,
		CredentialsFilePath: credentialsFile,
	}
	return cluster, provider, nil
}

func buildConfigForGardener(config model.GardenerConfig, credentialsFile string) (*types.Cluster, *types.Provider, error) {
	diskSize, err := strconv.Atoi(config.VolumeSize)

	if err != nil {
		return &types.Cluster{}, &types.Provider{}, err
	}

	cluster := &types.Cluster{
		KubernetesVersion: config.KubernetesVersion,
		Name:              config.Name,
		DiskSizeGB:        diskSize,
		NodeCount:         config.NodeCount,
		Location:          config.Region,
		MachineType:       config.MachineType,
	}

	provider := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         config.ProjectName,
		CredentialsFilePath: credentialsFile,
		CustomConfigurations: map[string]interface{}{
			"target_provider": config.TargetProvider,
			"target_secret":   config.TargetSecret,
			"disk_type":       config.DiskType,
			"zone":            config.Zone,
			"cidr":            config.Cidr,
			"autoscaler_min":  config.AutoScalerMin,
			"autoscaler_max":  config.AutoScalerMax,
			"max_surge":       config.MaxSurge,
			"max_unavailable": config.MaxUnavailable,
		},
	}
	return cluster, provider, nil
}
