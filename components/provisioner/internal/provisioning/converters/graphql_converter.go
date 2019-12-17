package converters

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type GraphQLConverter interface {
	RuntimeStatusToGraphQLStatus(status model.RuntimeStatus) *gqlschema.RuntimeStatus
	OperationStatusToGQLOperationStatus(operation model.Operation) *gqlschema.OperationStatus
}

func NewGraphQLConverter() GraphQLConverter {
	return &graphQLConverter{}
}

type graphQLConverter struct{}

func (c graphQLConverter) RuntimeStatusToGraphQLStatus(status model.RuntimeStatus) *gqlschema.RuntimeStatus {
	return &gqlschema.RuntimeStatus{
		LastOperationStatus:     c.OperationStatusToGQLOperationStatus(status.LastOperationStatus),
		RuntimeConnectionStatus: c.runtimeConnectionStatusToGraphQLStatus(status.RuntimeConnectionStatus),
		RuntimeConfiguration:    c.clusterToToGraphQLRuntimeConfiguration(status.RuntimeConfiguration),
	}
}

func (c graphQLConverter) OperationStatusToGQLOperationStatus(operation model.Operation) *gqlschema.OperationStatus {
	return &gqlschema.OperationStatus{
		ID:        &operation.ID,
		Operation: c.operationTypeToGraphQLType(operation.Type),
		State:     c.operationStateToGraphQLState(operation.State),
		Message:   &operation.Message,
		RuntimeID: &operation.ClusterID,
	}
}

func (c graphQLConverter) runtimeConnectionStatusToGraphQLStatus(status model.RuntimeAgentConnectionStatus) *gqlschema.RuntimeConnectionStatus {
	return &gqlschema.RuntimeConnectionStatus{Status: c.runtimeAgentConnectionStatusToGraphQLStatus(status)}
}

func (c graphQLConverter) runtimeAgentConnectionStatusToGraphQLStatus(status model.RuntimeAgentConnectionStatus) gqlschema.RuntimeAgentConnectionStatus {
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

func (c graphQLConverter) clusterToToGraphQLRuntimeConfiguration(config model.Cluster) *gqlschema.RuntimeConfig {
	return &gqlschema.RuntimeConfig{
		ClusterConfig:         c.clusterConfigToGraphQLConfig(config.ClusterConfig),
		KymaConfig:            c.kymaConfigToGraphQLConfig(config.KymaConfig),
		Kubeconfig:            config.Kubeconfig,
		CredentialsSecretName: &config.CredentialsSecretName,
	}
}

func (c graphQLConverter) clusterConfigToGraphQLConfig(config interface{}) gqlschema.ClusterConfig {
	gardenerConfig, ok := config.(model.GardenerConfig)
	if ok {
		return c.gardenerConfigToGraphQLConfig(gardenerConfig)
	}

	gcpConfig, ok := config.(model.GCPConfig)
	if ok {
		return c.gcpConfigToGraphQLConfig(gcpConfig)
	}
	return nil
}

func (c graphQLConverter) gardenerConfigToGraphQLConfig(config model.GardenerConfig) gqlschema.ClusterConfig {

	providerSpecificConfig := config.GardenerProviderConfig.AsProviderSpecificConfig()

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

func (c graphQLConverter) gcpConfigToGraphQLConfig(config model.GCPConfig) gqlschema.ClusterConfig {
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

func (c graphQLConverter) kymaConfigToGraphQLConfig(config model.KymaConfig) *gqlschema.KymaConfig {
	var components []*gqlschema.ComponentConfiguration
	for _, cmp := range config.Components {
		componentName := gqlschema.KymaComponent(cmp.Component)
		component := gqlschema.ComponentConfiguration{
			Component:     &componentName,
			Configuration: c.configurationToGraphQLConfig(cmp.Configuration),
		}

		components = append(components, &component)
	}

	return &gqlschema.KymaConfig{
		Version:       &config.Release.Version,
		Components:    components,
		Configuration: c.configurationToGraphQLConfig(config.GlobalConfiguration),
	}
}

func (c graphQLConverter) configurationToGraphQLConfig(cfg model.Configuration) []*gqlschema.ConfigEntry {
	configuration := make([]*gqlschema.ConfigEntry, 0, len(cfg.ConfigEntries))

	for _, configEntry := range cfg.ConfigEntries {
		secret := configEntry.Secret

		configuration = append(configuration, &gqlschema.ConfigEntry{
			Key:    configEntry.Key,
			Value:  configEntry.Value,
			Secret: &secret,
		})
	}

	return configuration
}

func (c graphQLConverter) operationTypeToGraphQLType(operationType model.OperationType) gqlschema.OperationType {
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

func (c graphQLConverter) operationStateToGraphQLState(state model.OperationState) gqlschema.OperationState {
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
