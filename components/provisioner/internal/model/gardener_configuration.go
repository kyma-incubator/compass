package model

import (
	"encoding/json"
	"errors"

	"github.com/kyma-incubator/hydroform/types"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type GardenerConfig struct {
	ID                     string
	ClusterID              string
	Name                   string
	ProjectName            string
	KubernetesVersion      string
	NodeCount              int
	VolumeSizeGB           int
	DiskType               string
	MachineType            string
	Provider               string
	Seed                   string
	TargetSecret           string
	Region                 string
	WorkerCidr             string
	AutoScalerMin          int
	AutoScalerMax          int
	MaxSurge               int
	MaxUnavailable         int
	GardenerProviderConfig GardenerProviderConfig
}

func (c GardenerConfig) ToHydroformConfiguration(credentialsFilePath string) (*types.Cluster, *types.Provider) {
	cluster := &types.Cluster{
		KubernetesVersion: c.KubernetesVersion,
		Name:              c.Name,
		DiskSizeGB:        c.VolumeSizeGB,
		NodeCount:         c.NodeCount,
		Location:          c.Region,
		MachineType:       c.MachineType,
	}

	customConfiguration := c.GardenerProviderConfig.AsMap()

	customConfiguration["target_provider"] = c.Provider
	customConfiguration["target_seed"] = c.Seed
	customConfiguration["target_secret"] = c.TargetSecret
	customConfiguration["disk_type"] = c.DiskType
	customConfiguration["workercidr"] = c.WorkerCidr
	customConfiguration["autoscaler_min"] = c.AutoScalerMin
	customConfiguration["autoscaler_max"] = c.AutoScalerMax
	customConfiguration["max_surge"] = c.MaxSurge
	customConfiguration["max_unavailable"] = c.MaxUnavailable

	provider := &types.Provider{
		Type:                 types.Gardener,
		ProjectName:          c.ProjectName,
		CredentialsFilePath:  credentialsFilePath,
		CustomConfigurations: customConfiguration,
	}

	return cluster, provider
}

type ProviderSpecificConfig string

func (c ProviderSpecificConfig) RawJSON() string {
	return string(c)
}

type GardenerProviderConfig interface {
	AsMap() map[string]interface{}
	RawJSON() string
}

type GCPGardenerConfig struct {
	ProviderSpecificConfig
	input gqlschema.GCPProviderConfigInput `db:"-"`
}

func NewGCPGardenerConfig(input gqlschema.GCPProviderConfigInput) (GCPGardenerConfig, error) {
	config, err := json.Marshal(input)
	if err != nil {
		return GCPGardenerConfig{}, errors.New("failed to marshal GCP Gardener config")
	}

	return GCPGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c GCPGardenerConfig) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"zone": c.input.Zone,
	}
}

type AzureGardenerConfig struct {
	ProviderSpecificConfig
	input gqlschema.AzureProviderConfigInput `db:"-"`
}

func NewAzureGardenerConfig(input gqlschema.AzureProviderConfigInput) (AzureGardenerConfig, error) {
	config, err := json.Marshal(input)
	if err != nil {
		return AzureGardenerConfig{}, errors.New("failed to marshal GCP Gardener config")
	}

	return AzureGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c AzureGardenerConfig) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"vnetcidr": c.input.VnetCidr,
	}
}

type AWSGardenerConfig struct {
	ProviderSpecificConfig
	input gqlschema.AWSProviderConfigInput `db:"-"`
}

func NewAWSGardenerConfig(input gqlschema.AWSProviderConfigInput) (AWSGardenerConfig, error) {
	config, err := json.Marshal(input)
	if err != nil {
		return AWSGardenerConfig{}, errors.New("failed to marshal GCP Gardener config")
	}

	return AWSGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c AWSGardenerConfig) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"zone":          c.input.Zone,
		"internalscidr": c.input.InternalCidr,
		"vpccidr":       c.input.VpcCidr,
		"publicscidr":   c.input.PublicCidr,
	}
}
