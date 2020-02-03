package provisioning

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/hyperscaler"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type InputConverter interface {
	ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput, tenant string) (model.Cluster, error)
}

func NewInputConverter(uuidGenerator uuid.UUIDGenerator,
	releaseRepo release.ReadRepository,
	gardenerProject string,
	hyperscalerAccountProvider hyperscaler.AccountProvider) InputConverter {
	return &converter{
		uuidGenerator:              uuidGenerator,
		releaseRepo:                releaseRepo,
		gardenerProject:            gardenerProject,
		hyperscalerAccountProvider: hyperscalerAccountProvider,
	}
}

type converter struct {
	uuidGenerator              uuid.UUIDGenerator
	releaseRepo                release.ReadRepository
	gardenerProject            string
	hyperscalerAccountProvider hyperscaler.AccountProvider
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
		providerConfig, err = c.providerConfigFromInput(runtimeID, *input.ClusterConfig, tenant)
		if err != nil {
			return model.Cluster{}, err
		}
	}

	credSecretName, err := c.hyperscalerAccountProvider.CompassSecretName(&input, tenant)

	if err != nil {
		return model.Cluster{}, err
	}

	return model.Cluster{
		ID:                    runtimeID,
		CredentialsSecretName: credSecretName,
		KymaConfig:            kymaConfig,
		ClusterConfig:         providerConfig,
		Tenant:                tenant,
	}, nil
}

func (c converter) providerConfigFromInput(runtimeID string, input gqlschema.ClusterConfigInput, tenant string) (model.ProviderConfiguration, error) {
	if input.GardenerConfig != nil {
		config := input.GardenerConfig
		return c.gardenerConfigFromInput(runtimeID, *config, tenant)
	}
	if input.GcpConfig != nil {
		config := input.GcpConfig
		return c.gcpConfigFromInput(runtimeID, *config), nil
	}
	return nil, errors.New("cluster config does not match any provider")
}

func (c converter) gardenerConfigFromInput(runtimeID string, input gqlschema.GardenerConfigInput, tenant string) (model.GardenerConfig, error) {
	id := c.uuidGenerator.New()
	name := c.createGardenerClusterName(input.Provider)

	providerSpecificConfig, err := c.providerSpecificConfigFromInput(input.ProviderSpecificConfig)

	if err != nil {
		return model.GardenerConfig{}, err
	}

	var seed string
	if input.Seed != nil {
		seed = *input.Seed
	}

	targetSecret, err := c.hyperscalerAccountProvider.GardenerSecretName(&input, tenant)

	if err != nil {
		return model.GardenerConfig{}, err
	}

	return model.GardenerConfig{
		ID:                     id,
		Name:                   name,
		ProjectName:            c.gardenerProject,
		KubernetesVersion:      input.KubernetesVersion,
		NodeCount:              input.NodeCount,
		VolumeSizeGB:           input.VolumeSizeGb,
		DiskType:               input.DiskType,
		MachineType:            input.MachineType,
		Provider:               input.Provider,
		Seed:                   seed,
		TargetSecret:           targetSecret,
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

func (c converter) createGardenerClusterName(provider string) string {
	id := c.uuidGenerator.New()
	provider = util.RemoveNotAllowedCharacters(provider)

	name := strings.ReplaceAll(id, "-", "")
	name = fmt.Sprintf("%.3s-%.7s", provider, name)
	name = util.StartWithLetter(name)
	name = strings.ToLower(name)
	return name
}

func HyperscalerTypeFromProviderInput(input *gqlschema.ProviderSpecificInput) (hyperscaler.HyperscalerType, error) {

	if input == nil {
		return hyperscaler.HyperscalerType(""), errors.New("ProviderSpecificInput not specified (was nil)")
	}

	if input.GcpConfig != nil {
		return hyperscaler.GCP, nil
	}
	if input.AzureConfig != nil {
		return hyperscaler.Azure, nil
	}
	if input.AwsConfig != nil {
		return hyperscaler.AWS, nil
	}

	return hyperscaler.HyperscalerType(""), errors.New("ProviderSpecificInput not specified")
}

func (c converter) providerSpecificConfigFromInput(input *gqlschema.ProviderSpecificInput) (model.GardenerProviderConfig, error) {

	hyperscalerType, err := HyperscalerTypeFromProviderInput(input)
	if err != nil {
		return nil, err
	}

	switch hyperscalerType {
	case hyperscaler.GCP:
		return model.NewGCPGardenerConfig(input.GcpConfig)

	case hyperscaler.Azure:
		return model.NewAzureGardenerConfig(input.AzureConfig)

	case hyperscaler.AWS:
		return model.NewAWSGardenerConfig(input.AwsConfig)

	}

	return nil, errors.Errorf("Unexpected HyperscalerType: %s from ProviderSpecificInput", hyperscalerType)
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
