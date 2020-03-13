package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	"github.com/kyma-incubator/hydroform/types"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
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

func (c GardenerConfig) ToShootTemplate(namespace string) (*gardener_types.Shoot, error) {
	allowPrivlagedContainers := true
	enableBasicAuthentication := false

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
			SecretBindingName: c.TargetSecret,
			SeedName:          seed,
			Region:            c.Region,
			Kubernetes: gardener_types.Kubernetes{
				AllowPrivilegedContainers: &allowPrivlagedContainers,
				Version:                   c.KubernetesVersion,
				KubeAPIServer: &gardener_types.KubeAPIServerConfig{
					EnableBasicAuthentication: &enableBasicAuthentication,
					AdmissionPlugins: []gardener_types.AdmissionPlugin{
						{Name: "SecurityContextDeny"},
					},
				},
			},
			Networking: gardener_types.Networking{
				Type:  "calico",        // Default value - we may consider adding it to API (if Hydroform will support it)
				Nodes: "10.250.0.0/19", // TODO: it is required - provide configuration in API (when Hydroform will support it)
			},
			Maintenance: &gardener_types.Maintenance{},
		},
	}

	err := c.GardenerProviderConfig.ExtendShootConfig(c, shoot)
	if err != nil {
		return nil, fmt.Errorf("error extending shoot config with Provider: %s", err.Error())
	}

	return shoot, nil
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
	ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) error
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

func (c GCPGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) error {
	shoot.Spec.CloudProfileName = "gcp"

	workers := make([]gardener_types.Worker, gardenerConfig.NodeCount)
	for i := 0; i < gardenerConfig.NodeCount; i++ {
		workers[i] = getWorkerConfig(gardenerConfig, []string{c.input.Zone}, i)
	}

	gcpInfra := NewGCPInfrastructure(gardenerConfig.WorkerCidr)
	jsonData, err := json.Marshal(gcpInfra)
	if err != nil {
		return fmt.Errorf("error encoding infrastructure config: %s", err.Error())
	}

	gcpControlPlane := NewGCPControlPlane(c.input.Zone)
	jsonCPData, err := json.Marshal(gcpControlPlane)
	if err != nil {
		return fmt.Errorf("error encoding control plane config: %s", err.Error())
	}

	shoot.Spec.Provider = gardener_types.Provider{
		Type:                 "gcp",
		ControlPlaneConfig:   &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonCPData}},
		InfrastructureConfig: &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonData}},
		Workers:              workers,
	}

	return nil
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

func (c AzureGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) error {
	shoot.Spec.CloudProfileName = "az"

	workers := make([]gardener_types.Worker, gardenerConfig.NodeCount)
	for i := 0; i < gardenerConfig.NodeCount; i++ {
		workers[i] = getWorkerConfig(gardenerConfig, nil, i)
	}

	azInfra := NewAzureInfrastructure(gardenerConfig.WorkerCidr, c)
	jsonData, err := json.Marshal(azInfra)
	if err != nil {
		return fmt.Errorf("error encoding infrastructure config: %s", err.Error())
	}

	azureControlPlane := NewAzureControlPlane()
	jsonCPData, err := json.Marshal(azureControlPlane)
	if err != nil {
		return fmt.Errorf("error encoding control plane config: %s", err.Error())
	}

	shoot.Spec.Provider = gardener_types.Provider{
		Type:                 "azure",
		ControlPlaneConfig:   &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonCPData}},
		InfrastructureConfig: &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonData}},
		Workers:              workers,
	}

	return nil
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

func (c AWSGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) error {
	shoot.Spec.CloudProfileName = "aws"

	workers := make([]gardener_types.Worker, gardenerConfig.NodeCount)
	for i := 0; i < gardenerConfig.NodeCount; i++ {
		workers[i] = getWorkerConfig(gardenerConfig, []string{c.input.Zone}, i)
	}

	awsInfra := NewAWSInfrastructure(gardenerConfig.WorkerCidr, c)
	jsonData, err := json.Marshal(awsInfra)
	if err != nil {
		return fmt.Errorf("error encoding infrastructure config: %s", err.Error())
	}

	awsControlPlane := NewAWSControlPlane()
	jsonCPData, err := json.Marshal(awsControlPlane)
	if err != nil {
		return fmt.Errorf("error encoding control plane config: %s", err.Error())
	}

	shoot.Spec.Provider = gardener_types.Provider{
		Type:                 "aws",
		ControlPlaneConfig:   &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonCPData}},
		InfrastructureConfig: &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonData}},
		Workers:              workers,
	}

	return nil
}

func getWorkerConfig(gardenerConfig GardenerConfig, zones []string, index int) gardener_types.Worker {
	return gardener_types.Worker{
		Name:           fmt.Sprintf("cpu-worker-%d", index),
		MaxSurge:       util.IntOrStrPtr(intstr.FromInt(gardenerConfig.MaxSurge)),
		MaxUnavailable: util.IntOrStrPtr(intstr.FromInt(gardenerConfig.MaxUnavailable)),
		Machine: gardener_types.Machine{
			Type: gardenerConfig.MachineType,
		},
		Volume: &gardener_types.Volume{
			Type: &gardenerConfig.DiskType,
			Size: fmt.Sprintf("%dGi", gardenerConfig.VolumeSizeGB),
		},
		Maximum: int32(gardenerConfig.AutoScalerMax),
		Minimum: int32(gardenerConfig.AutoScalerMin),
		Zones:   zones,
	}
}
