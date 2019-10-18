package provisioning

import (
	"github.com/gofrs/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

func RuntimeConfigFromInput(runtimeID string, input gqlschema.ProvisionRuntimeInput) (model.RuntimeConfig, error) {
	kymaConfig, err := kymaConfigFromInput(runtimeID, *input.KymaConfig)
	if err != nil {
		return model.RuntimeConfig{}, err
	}

	clusterConfig, err := clusterConfigFromInput(runtimeID, *input.ClusterConfig)
	if err != nil {
		return model.RuntimeConfig{}, err
	}

	return model.RuntimeConfig{
		KymaConfig:    kymaConfig,
		ClusterConfig: clusterConfig,
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
		ID:        operation.ID,
		Operation: operationTypeToGraphQLType(operation.Type),
		State:     operationStateToGraphQLState(operation.State),
		Message:   operation.Message,
		RuntimeID: operation.ClusterID,
	}
}

func clusterConfigFromInput(runtimeID string, input gqlschema.ClusterConfigInput) (interface{}, error) {
	if input.GardenerConfig != nil {
		config := input.GardenerConfig
		return gardenerConfigFromInput(runtimeID, *config)
	}
	if input.GcpConfig != nil {
		config := input.GcpConfig
		return gcpConfigFromInput(runtimeID, *config)
	}
	return nil, nil
}

func gardenerConfigFromInput(runtimeID string, input gqlschema.GardenerConfigInput) (model.GardenerConfig, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return model.GardenerConfig{}, dberrors.Internal("Failed to generate uuid for GardenerConfig: %s.", err)
	}

	return model.GardenerConfig{
		ID:                id.String(),
		Name:              input.Name,
		ProjectName:       input.ProjectName,
		KubernetesVersion: input.KubernetesVersion,
		NodeCount:         input.NodeCount,
		VolumeSize:        input.VolumeSize,
		DiskType:          input.DiskType,
		MachineType:       input.MachineType,
		TargetProvider:    input.TargetProvider,
		TargetSecret:      input.TargetSecret,
		Cidr:              input.Cidr,
		Region:            input.Region,
		Zone:              input.Zone,
		AutoScalerMin:     input.AutoScalerMin,
		AutoScalerMax:     input.AutoScalerMax,
		MaxSurge:          input.MaxSurge,
		MaxUnavailable:    input.MaxUnavailable,
		ClusterID:         runtimeID,
	}, nil
}

func gcpConfigFromInput(runtimeID string, input gqlschema.GCPConfigInput) (model.GCPConfig, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return model.GCPConfig{}, dberrors.Internal("Failed to generate uuid for GardenerConfig: %s.", err)
	}

	// TODO - Add unit test for this
	zone := ""
	if input.Zone != nil {
		zone = *input.Zone
	}

	return model.GCPConfig{
		ID:                id.String(),
		Name:              input.Name,
		ProjectName:       input.ProjectName,
		KubernetesVersion: input.KubernetesVersion,
		NumberOfNodes:     input.NumberOfNodes,
		BootDiskSize:      input.BootDiskSize,
		MachineType:       input.MachineType,
		Region:            input.Region,
		Zone:              zone,
		ClusterID:         runtimeID,
	}, nil
}

func kymaConfigFromInput(runtimeID string, input gqlschema.KymaConfigInput) (model.KymaConfig, error) {
	var modules []model.KymaConfigModule
	kymaConfigID, err := uuid.NewV4()
	if err != nil {
		return model.KymaConfig{}, dberrors.Internal("Failed to generate uuid for KymaConfig: %s.", err)
	}

	for _, module := range input.Modules {
		id, err := uuid.NewV4()
		if err != nil {
			return model.KymaConfig{}, dberrors.Internal("Failed to generate uuid for KymaConfigModule: %s.", err)
		}

		kymaConfigModule := model.KymaConfigModule{
			ID:           id.String(),
			Module:       model.KymaModule(module.String()),
			KymaConfigID: kymaConfigID.String(),
		}

		modules = append(modules, kymaConfigModule)
	}

	return model.KymaConfig{
		ID:        kymaConfigID.String(),
		Version:   input.Version,
		Modules:   modules,
		ClusterID: runtimeID,
	}, nil
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
		ClusterConfig: clusterConfigToGraphQLConfig(config.ClusterConfig),
		KymaConfig:    kymaConfigToGraphQLConfig(config.KymaConfig),
		Kubeconfig:    &config.Kubeconfig,
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

	return gqlschema.GardenerConfig{
		Name:              &config.Name,
		ProjectName:       &config.ProjectName,
		KubernetesVersion: &config.KubernetesVersion,
		NodeCount:         &config.NodeCount,
		DiskType:          &config.DiskType,
		VolumeSize:        &config.VolumeSize,
		MachineType:       &config.MachineType,
		TargetProvider:    &config.TargetProvider,
		TargetSecret:      &config.TargetSecret,
		Cidr:              &config.Cidr,
		Region:            &config.Region,
		Zone:              &config.Zone,
		AutoScalerMin:     &config.AutoScalerMin,
		AutoScalerMax:     &config.AutoScalerMax,
		MaxSurge:          &config.MaxSurge,
		MaxUnavailable:    &config.MaxUnavailable,
	}
}

func gcpConfigToGraphQLConfig(config model.GCPConfig) gqlschema.ClusterConfig {
	return gqlschema.GCPConfig{
		Name:              &config.Name,
		ProjectName:       &config.ProjectName,
		KubernetesVersion: &config.KubernetesVersion,
		NumberOfNodes:     &config.NumberOfNodes,
		BootDiskSize:      &config.BootDiskSize,
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
