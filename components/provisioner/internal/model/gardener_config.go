package model

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/apperrors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SubAccountLabel = "subaccount"
	AccountLabel    = "account"

	LicenceTypeAnnotation = "compass.provisioner.kyma-project.io/licence-type"
)

type GardenerConfig struct {
	ID                     string
	ClusterID              string
	Name                   string
	ProjectName            string
	KubernetesVersion      string
	VolumeSizeGB           int
	DiskType               string
	MachineType            string
	Provider               string
	Purpose                *string
	LicenceType            *string
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

func (c GardenerConfig) ToShootTemplate(namespace string, accountId string, subAccountId string) (*gardener_types.Shoot, apperrors.AppError) {
	allowPrivlagedContainers := true
	enableBasicAuthentication := false

	var seed *string = nil
	if c.Seed != "" {
		seed = util.StringPtr(c.Seed)
	}
	var purpose *gardener_types.ShootPurpose = nil
	if c.Purpose != nil && *c.Purpose != "" {
		p := gardener_types.ShootPurpose(*c.Purpose)
		purpose = &p
	}

	annotations := make(map[string]string)
	if c.LicenceType != nil {
		annotations[LicenceTypeAnnotation] = *c.LicenceType
	}

	shoot := &gardener_types.Shoot{
		ObjectMeta: v1.ObjectMeta{
			Name:      c.Name,
			Namespace: namespace,
			Labels: map[string]string{
				SubAccountLabel: subAccountId,
				AccountLabel:    accountId,
			},
			Annotations: annotations,
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
				},
			},
			Networking: gardener_types.Networking{
				Type:  "calico",                        // Default value - we may consider adding it to API (if Hydroform will support it)
				Nodes: util.StringPtr("10.250.0.0/19"), // TODO: it is required - provide configuration in API (when Hydroform will support it)
			},
			Purpose: purpose,
			Maintenance: &gardener_types.Maintenance{
				AutoUpdate: &gardener_types.MaintenanceAutoUpdate{
					KubernetesVersion:   false,
					MachineImageVersion: false,
				},
			},
		},
	}

	err := c.GardenerProviderConfig.ExtendShootConfig(c, shoot)
	if err != nil {
		return nil, err.Append("error extending shoot config with Provider")
	}

	return shoot, nil
}

type ProviderSpecificConfig string

func (c ProviderSpecificConfig) RawJSON() string {
	return string(c)
}

type GardenerProviderConfig interface {
	AsMap() (map[string]interface{}, apperrors.AppError)
	RawJSON() string
	AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig
	ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) apperrors.AppError
}

func NewGardenerProviderConfigFromJSON(jsonData string) (GardenerProviderConfig, apperrors.AppError) { //TODO: change to detect Provider correctly
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

	return nil, apperrors.BadRequest("json data does not match any of Gardener providers")
}

type GCPGardenerConfig struct {
	ProviderSpecificConfig
	input *gqlschema.GCPProviderConfigInput `db:"-"`
}

func NewGCPGardenerConfig(input *gqlschema.GCPProviderConfigInput) (*GCPGardenerConfig, apperrors.AppError) {
	config, err := json.Marshal(input)
	if err != nil {
		return &GCPGardenerConfig{}, apperrors.Internal("failed to marshal GCP Gardener config")
	}

	return &GCPGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c *GCPGardenerConfig) AsMap() (map[string]interface{}, apperrors.AppError) {
	if c.input == nil {
		err := json.Unmarshal([]byte(c.ProviderSpecificConfig), &c.input)
		if err != nil {
			return nil, apperrors.Internal("failed to decode Gardener GCP config: %s", err.Error())
		}
	}

	return map[string]interface{}{
		"zones": c.input.Zones,
	}, nil
}

func (c GCPGardenerConfig) AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig {
	return gqlschema.GCPProviderConfig{Zones: c.input.Zones}
}

func (c GCPGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) apperrors.AppError {
	shoot.Spec.CloudProfileName = "gcp"

	workers := []gardener_types.Worker{getWorkerConfig(gardenerConfig, c.input.Zones)}

	gcpInfra := NewGCPInfrastructure(gardenerConfig.WorkerCidr)
	jsonData, err := json.Marshal(gcpInfra)
	if err != nil {
		return apperrors.Internal("error encoding infrastructure config: %s", err.Error())
	}

	gcpControlPlane := NewGCPControlPlane(c.input.Zones)
	jsonCPData, err := json.Marshal(gcpControlPlane)
	if err != nil {
		return apperrors.Internal("error encoding control plane config: %s", err.Error())
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

func NewAzureGardenerConfig(input *gqlschema.AzureProviderConfigInput) (*AzureGardenerConfig, apperrors.AppError) {
	config, err := json.Marshal(input)
	if err != nil {
		return &AzureGardenerConfig{}, apperrors.Internal("failed to marshal GCP Gardener config")
	}

	return &AzureGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c *AzureGardenerConfig) AsMap() (map[string]interface{}, apperrors.AppError) {
	if c.input == nil {
		err := json.Unmarshal([]byte(c.ProviderSpecificConfig), &c.input)
		if err != nil {
			return nil, apperrors.Internal("failed to decode Gardener Azure config: %s", err.Error())
		}
	}

	cfg := map[string]interface{}{
		"vnetcidr": c.input.VnetCidr,
	}
	if c.input.Zones != nil {
		cfg["zones"] = c.input.Zones
	}

	return cfg, nil
}

func (c AzureGardenerConfig) AsProviderSpecificConfig() gqlschema.ProviderSpecificConfig {
	return gqlschema.AzureProviderConfig{VnetCidr: &c.input.VnetCidr, Zones: c.input.Zones}
}

type AWSGardenerConfig struct {
	ProviderSpecificConfig
	input *gqlschema.AWSProviderConfigInput `db:"-"`
}

func (c AzureGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) apperrors.AppError {
	shoot.Spec.CloudProfileName = "az"

	workers := []gardener_types.Worker{getWorkerConfig(gardenerConfig, c.input.Zones)}

	azInfra := NewAzureInfrastructure(gardenerConfig.WorkerCidr, c)
	jsonData, err := json.Marshal(azInfra)
	if err != nil {
		return apperrors.Internal("error encoding infrastructure config: %s", err.Error())
	}

	azureControlPlane := NewAzureControlPlane(c.input.Zones)
	jsonCPData, err := json.Marshal(azureControlPlane)
	if err != nil {
		return apperrors.Internal("error encoding control plane config: %s", err.Error())
	}

	shoot.Spec.Provider = gardener_types.Provider{
		Type:                 "azure",
		ControlPlaneConfig:   &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonCPData}},
		InfrastructureConfig: &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonData}},
		Workers:              workers,
	}

	return nil
}

func NewAWSGardenerConfig(input *gqlschema.AWSProviderConfigInput) (*AWSGardenerConfig, apperrors.AppError) {
	config, err := json.Marshal(input)
	if err != nil {
		return &AWSGardenerConfig{}, apperrors.Internal("failed to marshal GCP Gardener config")
	}

	return &AWSGardenerConfig{
		ProviderSpecificConfig: ProviderSpecificConfig(config),
		input:                  input,
	}, nil
}

func (c *AWSGardenerConfig) AsMap() (map[string]interface{}, apperrors.AppError) {
	if c.input == nil {
		err := json.Unmarshal([]byte(c.ProviderSpecificConfig), &c.input)
		if err != nil {
			return nil, apperrors.Internal("failed to decode Gardener AWS config: %s", err.Error())
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

func (c AWSGardenerConfig) ExtendShootConfig(gardenerConfig GardenerConfig, shoot *gardener_types.Shoot) apperrors.AppError {
	shoot.Spec.CloudProfileName = "aws"

	workers := []gardener_types.Worker{getWorkerConfig(gardenerConfig, []string{c.input.Zone})}

	awsInfra := NewAWSInfrastructure(gardenerConfig.WorkerCidr, c)
	jsonData, err := json.Marshal(awsInfra)
	if err != nil {
		return apperrors.Internal("error encoding infrastructure config: %s", err.Error())
	}

	awsControlPlane := NewAWSControlPlane()
	jsonCPData, err := json.Marshal(awsControlPlane)
	if err != nil {
		return apperrors.Internal("error encoding control plane config: %s", err.Error())
	}

	shoot.Spec.Provider = gardener_types.Provider{
		Type:                 "aws",
		ControlPlaneConfig:   &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonCPData}},
		InfrastructureConfig: &gardener_types.ProviderConfig{RawExtension: apimachineryRuntime.RawExtension{Raw: jsonData}},
		Workers:              workers,
	}

	return nil
}

func getWorkerConfig(gardenerConfig GardenerConfig, zones []string) gardener_types.Worker {
	return gardener_types.Worker{
		Name:           "cpu-worker-0",
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
