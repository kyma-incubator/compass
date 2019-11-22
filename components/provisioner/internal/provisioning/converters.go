package provisioning

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

func runtimeConfigFromInput(runtimeID string, input gqlschema.ProvisionRuntimeInput, uuidGenerator uuid.UUIDGenerator) (model.RuntimeConfig, error) {
	kymaConfig := kymaConfigFromInput(runtimeID, *input.KymaConfig, uuidGenerator)

	clusterConfig, err := clusterConfigFromInput(runtimeID, *input.ClusterConfig, uuidGenerator)

	if err != nil {
		return model.RuntimeConfig{}, err
	}

	return model.RuntimeConfig{
		KymaConfig:            kymaConfig,
		ClusterConfig:         clusterConfig,
		CredentialsSecretName: input.Credentials.SecretName,
	}, nil
}

func runtimeStatusToGraphQLStatus(status model.RuntimeStatus) *gqlschema.RuntimeStatus {
	return &gqlschema.RuntimeStatus{
		LastOperationStatus:     operationStatusToGQLOperationStatus(status.LastOperationStatus),
		RuntimeConnectionStatus: runtimeConnectionStatusToGraphQLStatus(status.RuntimeConnectionStatus),
		RuntimeConfiguration:    runtimeConfigurationToGraphQLConfiguration(status.RuntimeConfiguration),
	}
}

func operationStatusToGQLOperationStatus(operation model.Operation) *gqlschema.OperationStatus {
	return &gqlschema.OperationStatus{
		ID:        &operation.ID,
		Operation: operationTypeToGraphQLType(operation.Type),
		State:     operationStateToGraphQLState(operation.State),
		Message:   &operation.Message,
		RuntimeID: &operation.ClusterID,
	}
}

func clusterConfigFromInput(runtimeID string, input gqlschema.ClusterConfigInput, uuidGenerator uuid.UUIDGenerator) (interface{}, error) {
	if input.GardenerConfig != nil {
		config := input.GardenerConfig
		return gardenerConfigFromInput(runtimeID, *config, uuidGenerator)
	}
	if input.GcpConfig != nil {
		config := input.GcpConfig
		return gcpConfigFromInput(runtimeID, *config, uuidGenerator), nil
	}
	return nil, errors.New("cluster config does not match any provider")
}

func gardenerConfigFromInput(runtimeID string, input gqlschema.GardenerConfigInput, uuidGenerator uuid.UUIDGenerator) (model.GardenerConfig, error) {
	id := uuidGenerator.New()

	providerSpecificConfig, err := providerSpecificConfigFromInput(input.ProviderSpecificConfig)

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
		ProviderSpecificConfig: providerSpecificConfig,
	}, nil
}

func providerSpecificConfigFromInput(input *gqlschema.ProviderSpecificInput) (string, error) {
	var providerConfig interface{}

	if input.GcpConfig != nil {
		providerConfig = input.GcpConfig
	}
	if input.AzureConfig != nil {
		providerConfig = input.AzureConfig
	}
	if input.AwsConfig != nil {
		providerConfig = input.AwsConfig
	}

	providerConfigJson, err := json.Marshal(providerConfig)

	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to build provider specific config: %s", err.Error()))
	}

	return string(providerConfigJson), nil
}

func gcpConfigFromInput(runtimeID string, input gqlschema.GCPConfigInput, uuidGenerator uuid.UUIDGenerator) model.GCPConfig {
	id := uuidGenerator.New()

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

func kymaConfigFromInput(runtimeID string, input gqlschema.KymaConfigInput, uuidGenerator uuid.UUIDGenerator) model.KymaConfig {
	var modules []model.KymaConfigModule
	kymaConfigID := uuidGenerator.New()

	for _, module := range input.Modules {
		id := uuidGenerator.New()

		kymaConfigModule := model.KymaConfigModule{
			ID:           id,
			Module:       model.KymaModule(module.String()),
			KymaConfigID: kymaConfigID,
		}

		modules = append(modules, kymaConfigModule)
	}

	return model.KymaConfig{
		ID:        kymaConfigID,
		Version:   input.Version,
		Modules:   modules,
		ClusterID: runtimeID,
	}
}

func runtimeConnectionStatusToGraphQLStatus(status model.RuntimeAgentConnectionStatus) *gqlschema.RuntimeConnectionStatus {
	return &gqlschema.RuntimeConnectionStatus{Status: runtimeAgentConnectionStatusToGraphQLStatus(status)}
}

func runtimeAgentConnectionStatusToGraphQLStatus(status model.RuntimeAgentConnectionStatus) gqlschema.RuntimeAgentConnectionStatus {
	switch status {
	case model.RuntimeAgentConnectionStatusConnected:
		return gqlschema.RuntimeAgentConnectionStatusConnected
	case model.RuntimeAgentConnectionStatusDisconnected:
		return gqlschema.RuntimeAgentConnectionStatusDisconnected
	case model.RuntimeAgentConnectionStatusPending:
		return gqlschema.RuntimeAgentConnectionStatusPending
	default:
		return ""
	}
}

func runtimeConfigurationToGraphQLConfiguration(config model.RuntimeConfig) *gqlschema.RuntimeConfig {
	return &gqlschema.RuntimeConfig{
		ClusterConfig:         clusterConfigToGraphQLConfig(config.ClusterConfig),
		KymaConfig:            kymaConfigToGraphQLConfig(config.KymaConfig),
		Kubeconfig:            config.Kubeconfig,
		CredentialsSecretName: &config.CredentialsSecretName,
	}
}

func clusterConfigToGraphQLConfig(config interface{}) gqlschema.ClusterConfig {
	gardenerConfig, ok := config.(model.GardenerConfig)
	if ok {
		return gardenerConfigToGraphQLConfig(gardenerConfig)
	}

	gcpConfig, ok := config.(model.GCPConfig)
	if ok {
		return gcpConfigToGraphQLConfig(gcpConfig)
	}
	return nil
}

func gardenerConfigToGraphQLConfig(config model.GardenerConfig) gqlschema.ClusterConfig {

	providerSpecificConfig := providerSpecificConfigToGQLConfig(config.ProviderSpecificConfig)

	return gqlschema.GardenerConfig{
		Name:                   &config.Name,
		ProjectName:            &config.ProjectName,
		KubernetesVersion:      &config.KubernetesVersion,
		NodeCount:              &config.NodeCount,
		DiskType:               &config.DiskType,
		VolumeSizeGb:           &config.VolumeSizeGB,
		MachineType:            &config.MachineType,
		Provider:               &config.Provider,
		Seed:                   &config.Seed,
		TargetSecret:           &config.TargetSecret,
		WorkerCidr:             &config.WorkerCidr,
		Region:                 &config.Region,
		AutoScalerMin:          &config.AutoScalerMin,
		AutoScalerMax:          &config.AutoScalerMax,
		MaxSurge:               &config.MaxSurge,
		MaxUnavailable:         &config.MaxUnavailable,
		ProviderSpecificConfig: providerSpecificConfig,
	}
}

func providerSpecificConfigToGQLConfig(config string) gqlschema.ProviderSpecificConfig {
	var gcpProviderConfig gqlschema.GCPProviderConfig
	err := util.DecodeJson(config, &gcpProviderConfig)
	if err == nil {
		return gcpProviderConfig
	}

	var azureProviderConfig gqlschema.AzureProviderConfig
	err = util.DecodeJson(config, &azureProviderConfig)
	if err == nil {
		return azureProviderConfig
	}

	var awsProviderConfig gqlschema.AWSProviderConfig
	err = util.DecodeJson(config, &awsProviderConfig)
	if err == nil {
		return awsProviderConfig
	}
	return nil
}

func gcpConfigToGraphQLConfig(config model.GCPConfig) gqlschema.ClusterConfig {
	return gqlschema.GCPConfig{
		Name:              &config.Name,
		ProjectName:       &config.ProjectName,
		KubernetesVersion: &config.KubernetesVersion,
		NumberOfNodes:     &config.NumberOfNodes,
		BootDiskSizeGb:    &config.BootDiskSizeGB,
		MachineType:       &config.MachineType,
		Region:            &config.Region,
		Zone:              &config.Zone,
	}
}

func kymaConfigToGraphQLConfig(config model.KymaConfig) *gqlschema.KymaConfig {
	var modules []*gqlschema.KymaModule
	for _, module := range config.Modules {
		kymaModule := gqlschema.KymaModule(module.Module)
		modules = append(modules, &kymaModule)
	}

	return &gqlschema.KymaConfig{
		Version: &config.Version,
		Modules: modules,
	}
}

func operationTypeToGraphQLType(operationType model.OperationType) gqlschema.OperationType {
	switch operationType {
	case model.Provision:
		return gqlschema.OperationTypeProvision
	case model.Deprovision:
		return gqlschema.OperationTypeDeprovision
	case model.Upgrade:
		return gqlschema.OperationTypeUpgrade
	case model.ReconnectRuntime:
		return gqlschema.OperationTypeReconnectRuntime
	default:
		return ""
	}
}

func operationStateToGraphQLState(state model.OperationState) gqlschema.OperationState {
	switch state {
	case model.InProgress:
		return gqlschema.OperationStateInProgress
	case model.Succeeded:
		return gqlschema.OperationStateSucceeded
	case model.Failed:
		return gqlschema.OperationStateFailed
	default:
		return ""
	}
}
