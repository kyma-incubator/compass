package provisioning

import (
	"fmt"
	"strings"

	"github.com/kyma-project/control-plane/components/provisioner/internal/apperrors"

	"github.com/kyma-project/control-plane/components/provisioner/internal/installation/release"
	"github.com/kyma-project/control-plane/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-project/control-plane/components/provisioner/internal/util"

	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/uuid"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
)

type InputConverter interface {
	ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput, tenant, subAccountId string) (model.Cluster, apperrors.AppError)
	KymaConfigFromInput(runtimeID string, input gqlschema.KymaConfigInput) (model.KymaConfig, apperrors.AppError)
}

func NewInputConverter(uuidGenerator uuid.UUIDGenerator, releaseRepo release.Provider, gardenerProject string) InputConverter {
	return &converter{
		uuidGenerator:   uuidGenerator,
		releaseRepo:     releaseRepo,
		gardenerProject: gardenerProject,
	}
}

type converter struct {
	uuidGenerator   uuid.UUIDGenerator
	releaseRepo     release.Provider
	gardenerProject string
}

func (c converter) ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput, tenant, subAccountId string) (model.Cluster, apperrors.AppError) {
	var err apperrors.AppError

	var kymaConfig model.KymaConfig
	if input.KymaConfig != nil {
		kymaConfig, err = c.KymaConfigFromInput(runtimeID, *input.KymaConfig)
		if err != nil {
			return model.Cluster{}, err
		}
	}

	if input.ClusterConfig == nil {
		return model.Cluster{}, apperrors.BadRequest("error: ClusterConfig not provided")
	}

	gardenerConfig, err := c.gardenerConfigFromInput(runtimeID, input.ClusterConfig.GardenerConfig)
	if err != nil {
		return model.Cluster{}, err
	}

	return model.Cluster{
		ID:            runtimeID,
		KymaConfig:    kymaConfig,
		ClusterConfig: gardenerConfig,
		Tenant:        tenant,
		SubAccountId:  &subAccountId,
	}, nil
}

func (c converter) gardenerConfigFromInput(runtimeID string, input *gqlschema.GardenerConfigInput) (model.GardenerConfig, apperrors.AppError) {
	if input == nil {
		return model.GardenerConfig{}, apperrors.BadRequest("error: GardenerConfig not provided")
	}

	providerSpecificConfig, err := c.providerSpecificConfigFromInput(input.ProviderSpecificConfig)
	if err != nil {
		return model.GardenerConfig{}, err
	}

	return model.GardenerConfig{
		ID:                     c.uuidGenerator.New(),
		Name:                   c.createGardenerClusterName(),
		ProjectName:            c.gardenerProject,
		KubernetesVersion:      input.KubernetesVersion,
		VolumeSizeGB:           input.VolumeSizeGb,
		DiskType:               input.DiskType,
		MachineType:            input.MachineType,
		Provider:               input.Provider,
		Purpose:                input.Purpose,
		LicenceType:            input.LicenceType,
		Seed:                   util.UnwrapStr(input.Seed),
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

func (c converter) createGardenerClusterName() string {
	id := c.uuidGenerator.New()

	name := strings.ReplaceAll(id, "-", "")
	name = fmt.Sprintf("%.7s", name)
	name = util.StartWithLetter(name)
	name = strings.ToLower(name)
	return name
}

func (c converter) providerSpecificConfigFromInput(input *gqlschema.ProviderSpecificInput) (model.GardenerProviderConfig, apperrors.AppError) {
	if input == nil {
		return nil, apperrors.Internal("provider config not specified")
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

	return nil, apperrors.BadRequest("provider config not specified")
}

func (c converter) KymaConfigFromInput(runtimeID string, input gqlschema.KymaConfigInput) (model.KymaConfig, apperrors.AppError) {
	kymaRelease, err := c.releaseRepo.GetReleaseByVersion(input.Version)
	if err != nil {
		if err.Code() == dberrors.CodeNotFound {
			return model.KymaConfig{}, apperrors.BadRequest("Kyma Release %s not found", input.Version)
		}

		return model.KymaConfig{}, apperrors.Internal("Failed to get Kyma Release with version %s: %s", input.Version, err.Error())
	}

	var components []model.KymaComponentConfig
	kymaConfigID := c.uuidGenerator.New()

	for i, component := range input.Components {
		id := c.uuidGenerator.New()

		kymaConfigModule := model.KymaComponentConfig{
			ID:             id,
			Component:      model.KymaComponent(component.Component),
			Namespace:      component.Namespace,
			SourceURL:      component.SourceURL,
			Configuration:  c.configurationFromInput(component.Configuration),
			ComponentOrder: i + 1,
			KymaConfigID:   kymaConfigID,
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
