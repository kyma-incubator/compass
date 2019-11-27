package converters

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type InputConverter interface {
	ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput) (model.Cluster, error)
}

func NewInputConverter(uuidGenerator uuid.UUIDGenerator, session dbsession.ReadSession) InputConverter {
	return &converter{
		uuidGenerator: uuidGenerator,
		readSession:   session,
	}
}

type converter struct {
	uuidGenerator uuid.UUIDGenerator
	readSession   dbsession.ReadSession
}

func (c converter) ProvisioningInputToCluster(runtimeID string, input gqlschema.ProvisionRuntimeInput) (model.Cluster, error) {
	kymaConfig, err := c.kymaConfigFromInput(runtimeID, *input.KymaConfig)
	if err != nil {
		return model.Cluster{}, err
	}

	clusterConfig, err := c.clusterConfigFromInput(runtimeID, *input.ClusterConfig)
	if err != nil {
		return model.Cluster{}, err
	}

	return model.Cluster{
		KymaConfig:            kymaConfig,
		ClusterConfig:         clusterConfig,
		CredentialsSecretName: input.Credentials.SecretName,
	}, nil
}

func (c converter) clusterConfigFromInput(runtimeID string, input gqlschema.ClusterConfigInput) (model.ProviderConfiguration, error) {
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
	if input.GcpConfig != nil {
		return model.NewGCPGardenerConfig(*input.GcpConfig)
	}
	if input.AzureConfig != nil {
		return model.NewAzureGardenerConfig(*input.AzureConfig)
	}
	if input.AwsConfig != nil {
		return model.NewAWSGardenerConfig(*input.AwsConfig)
	}

	return nil, errors.New("provider config not provided")
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
	release, err := c.readSession.GetReleaseByVersion(input.Version)
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
		Release:   release,
		Modules:   modules,
		ClusterID: runtimeID,
	}, nil
}
