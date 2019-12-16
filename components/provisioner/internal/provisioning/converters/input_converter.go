package converters

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/hyperscaler"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type InputConverter interface {
	ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput) (model.Cluster, error)
}

func NewInputConverter(
	uuidGenerator uuid.UUIDGenerator,
	releaseRepo release.ReadRepository,
	hyperscalerAccountProvider hyperscaler.AccountProvider) InputConverter {

	return &converter{
		uuidGenerator:              uuidGenerator,
		releaseRepo:                releaseRepo,
		hyperscalerAccountProvider: hyperscalerAccountProvider,
	}
}

type converter struct {
	uuidGenerator              uuid.UUIDGenerator
	releaseRepo                release.ReadRepository
	hyperscalerAccountProvider hyperscaler.AccountProvider
}

func (c converter) ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput) (model.Cluster, error) {
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

	credSecretName, err := c.hyperscalerAccountProvider.CompassSecretName(&input)

	if err != nil {
		return model.Cluster{}, err
	}

	return model.Cluster{
		ID:                    runtimeID,
		CredentialsSecretName: credSecretName,
		KymaConfig:            kymaConfig,
		ClusterConfig:         providerConfig,
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

	targetSecret, err := c.hyperscalerAccountProvider.GardenerSecretName(&input)

	if err != nil {
		return model.GardenerConfig{}, err
	}

	gardenerConfig := model.GardenerConfig{
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
		TargetSecret:           targetSecret,
		WorkerCidr:             input.WorkerCidr,
		Region:                 input.Region,
		AutoScalerMin:          input.AutoScalerMin,
		AutoScalerMax:          input.AutoScalerMax,
		MaxSurge:               input.MaxSurge,
		MaxUnavailable:         input.MaxUnavailable,
		ClusterID:              runtimeID,
		GardenerProviderConfig: providerSpecificConfig,
	}

	return gardenerConfig, nil
}

func HyperscalerTypeFromProviderInput(input *gqlschema.ProviderSpecificInput) (hyperscaler.HyperscalerType, error) {

	if input == nil {
		return hyperscaler.HyperscalerType(""), errors.New("ProviderSpecificInput not specified (nil)")
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

	var modules []model.KymaConfigModule
	kymaConfigID := c.uuidGenerator.New()

	for _, module := range input.Modules {
		id := c.uuidGenerator.New()

		kymaConfigModule := model.KymaConfigModule{
			ID:           id,
			Module:       model.KymaModule(module.String()),
			KymaConfigID: kymaConfigID,
		}

		modules = append(modules, kymaConfigModule)
	}

	return model.KymaConfig{
		ID:        kymaConfigID,
		Release:   kymaRelease,
		Modules:   modules,
		ClusterID: runtimeID,
	}, nil
}
