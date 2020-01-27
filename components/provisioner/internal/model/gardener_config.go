package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"

	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

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

// TODO: test
func (c GardenerConfig) ToShootTemplate(namespace string) *gardener_types.Shoot {
	allowPrivlagedContainers := true

	var seed *string = nil
	if c.Seed != "" {
		seed = util.StringPtr(c.Seed)
	}

	shoot := &gardener_types.Shoot{
		ObjectMeta: v1.ObjectMeta{
			Name:      c.Name,
			Namespace: namespace,
		},
		Spec: gardener_types.ShootSpec{
			Cloud: gardener_types.Cloud{
				Region: c.Region,
				SecretBindingRef: corev1.LocalObjectReference{
					Name: c.TargetSecret,
				},
				Seed: seed,
			},
			Kubernetes: gardener_types.Kubernetes{
				AllowPrivilegedContainers: &allowPrivlagedContainers,
				Version:                   c.KubernetesVersion,
			},
		},
	}

	c.GardenerProviderConfig.ExtendShootConfig(c, shoot)

	return shoot
}

func (c GardenerConfig) ToHydroformConfiguration(credentialsFilePath string) (*types.Cluster, *types.Provider, error) {
	cluster := &types.Cluster{
		KubernetesVersion: c.KubernetesVersion,
		Name:              c.Name,
		DiskSizeGB:        c.VolumeSizeGB,
		NodeCount:         c.NodeCount,
		Location:          c.Region,
		MachineType:       c.MachineType,
	}

	customConfiguration, err := c.GardenerProviderConfig.AsMap()
	if err != nil {
		return nil, nil, err
	}

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

	return cluster, provider, nil
}

type ProviderSpecificConfig string

func (c ProviderSpecificConfig) RawJSON() string {
	return string(c)
}

type GardenerProviderConfig interface {
	AsMap() (map[string]interface{}, error)
	RawJSON() string
	AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig
	ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot)
}

func NewGardenerProviderConfigFromJSON(jsonData string) (GardenerProviderConfig, error) {
	var gcpProviderConfig gqlschema.GCPProviderConfigInput
	err := util.DecodeJson(jsonData, &gcpProviderConfig)
	if err == nil {
		return &GCPGardenerConfig{input: &gcpProviderConfig, ProviderSpecificConfig: ProviderSpecificConfig(jsonData)}, nil
	}

	var azureProviderConfig gqlschema.AzureProviderConfigInput
	err = util.DecodeJson(jsonData, &azureProviderConfig)
	if err == nil {
		return &AzureGardenerConfig{input: &azureProviderConfig, ProviderSpecificConfig: ProviderSpecificConfig(jsonData)}, nil
	}

	var awsProviderConfig gqlschema.AWSProviderConfigInput
	err = util.DecodeJson(jsonData, &awsProviderConfig)
	if err == nil {
		return &AWSGardenerConfig{input: &awsProviderConfig, ProviderSpecificConfig: ProviderSpecificConfig(jsonData)}, nil
	}

	return nil, errors.New("json data does not match any of Gardener providers")
}

type GCPGardenerConfig struct {
	ProviderSpecificConfig
	input *gqlschema.GCPProviderConfigInput `db:"-"`
}

func NewGCPGardenerConfig(input *gqlschema.GCPProviderConfigInput) (*GCPGardenerConfig, error) {
	config, err := json.Marshal(input)
	if err != nil {
		return &GCPGardenerConfig{}, errors.New("failed to marshal GCP Gardener config")
	}

	return &GCPGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c *GCPGardenerConfig) AsMap() (map[string]interface{}, error) {
	if c.input == nil {
		err := json.Unmarshal([]byte(c.ProviderSpecificConfig), &c.input)
		if err != nil {
			return nil, fmt.Errorf("failed to decode Gardener GCP config: %s", err.Error())
		}
	}

	return map[string]interface{}{
		"zone": c.input.Zone,
	}, nil
}

func (c GCPGardenerConfig) AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig {
	return gqlschema.GCPProviderConfig{Zone: &c.input.Zone}
}

func (c GCPGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) {
	shoot.Spec.Cloud.Profile = "gcp"

	workers := make([]gardener_types.GCPWorker, gardenerConfig.NodeCount)
	for i := 0; i < gardenerConfig.NodeCount; i++ {
		workers[i] = gardener_types.GCPWorker{
			Worker:     getWorkerConfig(gardenerConfig, i),
			VolumeType: gardenerConfig.DiskType,
			VolumeSize: fmt.Sprintf("%dGi", gardenerConfig.VolumeSizeGB),
		}
	}

	shoot.Spec.Cloud.GCP = &gardener_types.GCPCloud{
		Networks: gardener_types.GCPNetworks{
			K8SNetworks: gardener_types.K8SNetworks{},
			VPC:         nil,
			Workers:     []string{gardenerConfig.WorkerCidr},
			Internal:    nil,
		},
		Workers: workers,
		Zones:   []string{c.input.Zone},
	}
}

type AzureGardenerConfig struct {
	ProviderSpecificConfig
	input *gqlschema.AzureProviderConfigInput `db:"-"`
}

func NewAzureGardenerConfig(input *gqlschema.AzureProviderConfigInput) (*AzureGardenerConfig, error) {
	config, err := json.Marshal(input)
	if err != nil {
		return &AzureGardenerConfig{}, errors.New("failed to marshal GCP Gardener config")
	}

	return &AzureGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c *AzureGardenerConfig) AsMap() (map[string]interface{}, error) {
	if c.input == nil {
		err := json.Unmarshal([]byte(c.ProviderSpecificConfig), &c.input)
		if err != nil {
			return nil, fmt.Errorf("failed to decode Gardener Azure config: %s", err.Error())
		}
	}

	return map[string]interface{}{
		"vnetcidr": c.input.VnetCidr,
	}, nil
}

func (c AzureGardenerConfig) AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig {
	return gqlschema.AzureProviderConfig{VnetCidr: &c.input.VnetCidr}
}

type AWSGardenerConfig struct {
	ProviderSpecificConfig
	input *gqlschema.AWSProviderConfigInput `db:"-"`
}

func (c AzureGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) {
	shoot.Spec.Cloud.Profile = "az"

	workers := make([]gardener_types.AzureWorker, gardenerConfig.NodeCount)
	for i := 0; i < gardenerConfig.NodeCount; i++ {
		workers[i] = gardener_types.AzureWorker{
			Worker:     getWorkerConfig(gardenerConfig, i),
			VolumeType: gardenerConfig.DiskType,
			VolumeSize: fmt.Sprintf("%dGi", gardenerConfig.VolumeSizeGB),
		}
	}

	shoot.Spec.Cloud.Azure = &gardener_types.AzureCloud{
		Networks: gardener_types.AzureNetworks{
			K8SNetworks: gardener_types.K8SNetworks{},
			VNet:        gardener_types.AzureVNet{CIDR: &c.input.VnetCidr},
			Workers:     gardenerConfig.WorkerCidr,
		},
		Workers: workers,
	}
}

func NewAWSGardenerConfig(input *gqlschema.AWSProviderConfigInput) (*AWSGardenerConfig, error) {
	config, err := json.Marshal(input)
	if err != nil {
		return &AWSGardenerConfig{}, errors.New("failed to marshal GCP Gardener config")
	}

	return &AWSGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c *AWSGardenerConfig) AsMap() (map[string]interface{}, error) {
	if c.input == nil {
		err := json.Unmarshal([]byte(c.ProviderSpecificConfig), &c.input)
		if err != nil {
			return nil, fmt.Errorf("failed to decode Gardener AWS config: %s", err.Error())
		}
	}

	return map[string]interface{}{
		"zone":          c.input.Zone,
		"internalscidr": c.input.InternalCidr,
		"vpccidr":       c.input.VpcCidr,
		"publicscidr":   c.input.PublicCidr,
	}, nil
}

func (c AWSGardenerConfig) AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig {
	return gqlschema.AWSProviderConfig{
		Zone:         &c.input.Zone,
		VpcCidr:      &c.input.VpcCidr,
		PublicCidr:   &c.input.PublicCidr,
		InternalCidr: &c.input.InternalCidr,
	}
}

func (c AWSGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) {
	shoot.Spec.Cloud.Profile = "aws"

	workers := make([]gardener_types.AWSWorker, gardenerConfig.NodeCount)
	for i := 0; i < gardenerConfig.NodeCount; i++ {
		workers[i] = gardener_types.AWSWorker{
			Worker:     getWorkerConfig(gardenerConfig, i),
			VolumeType: gardenerConfig.DiskType,
			VolumeSize: fmt.Sprintf("%dGi", gardenerConfig.VolumeSizeGB),
		}
	}

	shoot.Spec.Cloud.AWS = &gardener_types.AWSCloud{
		Networks: gardener_types.AWSNetworks{
			K8SNetworks: gardener_types.K8SNetworks{},
			VPC:         gardener_types.AWSVPC{CIDR: util.StringPtr(c.input.VpcCidr)},
			Internal:    []string{c.input.InternalCidr},
			Public:      []string{c.input.PublicCidr},
			Workers:     []string{gardenerConfig.WorkerCidr},
		},
		Workers: workers,
		Zones:   []string{c.input.Zone},
	}
}

func getWorkerConfig(gardenerConfig GardenerConfig, index int) gardener_types.Worker {
	return gardener_types.Worker{
		Name:           fmt.Sprintf("cpu-worker-%d", index),
		MachineType:    gardenerConfig.MachineType,
		AutoScalerMin:  gardenerConfig.AutoScalerMin,
		AutoScalerMax:  gardenerConfig.AutoScalerMax,
		MaxSurge:       util.IntOrStrPtr(intstr.FromInt(gardenerConfig.MaxSurge)),
		MaxUnavailable: util.IntOrStrPtr(intstr.FromInt(gardenerConfig.MaxUnavailable)),
	}
}
