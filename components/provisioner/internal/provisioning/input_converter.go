package provisioning

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type InputConverter interface {
	ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput, tenant string) (model.Cluster, error)
}

func NewInputConverter(uuidGenerator uuid.UUIDGenerator, releaseRepo release.ReadRepository) InputConverter {
	return &converter{
		uuidGenerator: uuidGenerator,
		releaseRepo:   releaseRepo,
	}
}

type converter struct {
	uuidGenerator uuid.UUIDGenerator
	releaseRepo   release.ReadRepository
}

func (c converter) ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput, tenant string) (model.Cluster, error) {
	var err error

	var kymaConfig model.KymaConfig
	if input.KymaConfig != nil {
		kymaConfig, err = c.kymaConfigFromInput(runtimeID, *input.KymaConfig)
		if err != nil {
			return model.Cluster{}, err
		}
	}

	var providerConfig model.ProviderConfiguration
	if input.ClusterConfig != nil {
		providerConfig, err = c.providerConfigFromInput(runtimeID, *input.ClusterConfig)
		if err != nil {
			return model.Cluster{}, err
		}
	}

	var credSecretName string
	if input.Credentials != nil {
		credSecretName = input.Credentials.SecretName
	}

	return model.Cluster{
		ID:                    runtimeID,
		CredentialsSecretName: credSecretName,
		KymaConfig:            kymaConfig,
		ClusterConfig:         providerConfig,
		Tenant:                tenant,
	}, nil
}

func (c converter) providerConfigFromInput(runtimeID string, input gqlschema.ClusterConfigInput) (model.ProviderConfiguration, error) {
	if input.GardenerConfig != nil {
		config := input.GardenerConfig
		return c.gardenerConfigFromInput(runtimeID, *config)
	}
	if input.GcpConfig != nil {
		config := input.GcpConfig
		return c.gcpConfigFromInput(runtimeID, *config), nil
	}
	return nil, errors.New("cluster config does not match any provider")
}

func (c converter) gardenerConfigFromInput(runtimeID string, input gqlschema.GardenerConfigInput) (model.GardenerConfig, error) {
	id := c.uuidGenerator.New()

	providerSpecificConfig, err := c.providerSpecificConfigFromInput(input.ProviderSpecificConfig)

	if err != nil {
		return model.GardenerConfig{}, err
	}

	return model.GardenerConfig{
		ID:                     id,
		Name:                   input.Name,
		ProjectName:            input.ProjectName,
		KubernetesVersion:      input.KubernetesVersion,
		NodeCount:              input.NodeCount,
		VolumeSizeGB:           input.VolumeSizeGb,
		DiskType:               input.DiskType,
		MachineType:            input.MachineType,
		Provider:               input.Provider,
		Seed:                   input.Seed,
		TargetSecret:           input.TargetSecret,
		WorkerCidr:             input.WorkerCidr,
		Region:                 input.Region,
		AutoScalerMin:          input.AutoScalerMin,
		AutoScalerMax:          input.AutoScalerMax,
		MaxSurge:               input.MaxSurge,
		MaxUnavailable:         input.MaxUnavailable,
		ClusterID:              runtimeID,
		GardenerProviderConfig: providerSpecificConfig,
	}, nil
}

func (c converter) providerSpecificConfigFromInput(input *gqlschema.ProviderSpecificInput) (model.GardenerProviderConfig, error) {
	if input == nil {
		return nil, errors.New("provider config not specified")
	}

	if input.GcpConfig != nil {
		return model.NewGCPGardenerConfig(input.GcpConfig)
	}
	if input.AzureConfig != nil {
		return model.NewAzureGardenerConfig(input.AzureConfig)
	}
	if input.AwsConfig != nil {
		return model.NewAWSGardenerConfig(input.AwsConfig)
	}

	return nil, errors.New("provider config not specified")
}

func (c converter) gcpConfigFromInput(runtimeID string, input gqlschema.GCPConfigInput) model.GCPConfig {
	id := c.uuidGenerator.New()

	zone := ""
	if input.Zone != nil {
		zone = *input.Zone
	}

	return model.GCPConfig{
		ID:                id,
		Name:              input.Name,
		ProjectName:       input.ProjectName,
		KubernetesVersion: input.KubernetesVersion,
		NumberOfNodes:     input.NumberOfNodes,
		BootDiskSizeGB:    input.BootDiskSizeGb,
		MachineType:       input.MachineType,
		Region:            input.Region,
		Zone:              zone,
		ClusterID:         runtimeID,
	}
}

func (c converter) kymaConfigFromInput(runtimeID string, input gqlschema.KymaConfigInput) (model.KymaConfig, error) {
	kymaRelease, err := c.releaseRepo.GetReleaseByVersion(input.Version)
	if err != nil {
		if err.Code() == dberrors.CodeNotFound {
			return model.KymaConfig{}, errors.Errorf("Kyma Release %s not found", input.Version)
		}

		return model.KymaConfig{}, errors.WithMessagef(err, "Failed to get Kyma Release with version %s", input.Version)
	}

	var components []model.KymaComponentConfig
	kymaConfigID := c.uuidGenerator.New()

	for _, component := range input.Components {
		id := c.uuidGenerator.New()

		kymaConfigModule := model.KymaComponentConfig{
			ID:            id,
			Component:     model.KymaComponent(component.Component),
			Namespace:     component.Namespace,
			Configuration: c.configurationFromInput(component.Configuration),
			KymaConfigID:  kymaConfigID,
		}

		components = append(components, kymaConfigModule)
	}

	return model.KymaConfig{
		ID:                  kymaConfigID,
		Release:             kymaRelease,
		Components:          components,
		ClusterID:           runtimeID,
		GlobalConfiguration: c.configurationFromInput(input.Configuration),
	}, nil
}

func (c converter) configurationFromInput(input []*gqlschema.ConfigEntryInput) model.Configuration {
	configuration := model.Configuration{
		ConfigEntries: make([]model.ConfigEntry, 0, len(input)),
	}

	for _, ce := range input {
		configuration.ConfigEntries = append(configuration.ConfigEntries, configEntryFromInput(ce))
	}

	return configuration
}

func configEntryFromInput(entry *gqlschema.ConfigEntryInput) model.ConfigEntry {
	return model.NewConfigEntry(entry.Key, entry.Value, util.BoolFromPtr(entry.Secret))
}
