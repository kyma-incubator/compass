package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type InfrastructureProvider int

const (
	GCP      InfrastructureProvider = iota
	AKS      InfrastructureProvider = iota
	Gardener InfrastructureProvider = iota
)

type KymaModule string

type KymaConfig struct {
	Version string
	Modules []KymaModule
}

type OperationState string

const (
	InProgress OperationState = "IN_PROGRESS"
	Succeeded  OperationState = "SUCCEEDED"
	Failed     OperationState = "FAILED"
)

type OperationType string

const (
	Provision        OperationType = "PROVISION"
	Upgrade          OperationType = "UPGRADE"
	Deprovision      OperationType = "DEPROVISION"
	ReconnectRuntime OperationType = "RECONNECT_RUNTIME"
)

type Operation struct {
	OperationID string
	Operation   OperationType
	Started     time.Time
	Finished    time.Time
	State       OperationState
	Message     string
	RuntimeID   string
}

type GardenerConfig struct {
	Name              string
	ProjectName       string
	KubernetesVersion string
	NodeCount         int
	VolumeSize        string
	DiskType          string
	MachineType       string
	TargetProvider    string
	TargetSecret      string
	Cidr              string
	Region            string
	Zone              string
	AutoScalerMin     int
	AutoScalerMax     int
	MaxSurge          int
	MaxUnavailable    int
}

type GCPConfig struct {
	Name              string
	ProjectName       string
	KubernetesVersion string
	NumberOfNodes     int
	BootDiskSize      string
	MachineType       string
	Region            string
	Zone              string
}

type RuntimeAgentConnectionStatus int

type ClusterConfig struct {
	Name                   string
	NodeCount              int
	DiskSize               string
	MachineType            string
	Region                 string
	Version                string
	Credentials            string
	InfrastructureProvider InfrastructureProvider
	ProviderConfig         interface{}
}

const (
	RuntimeAgentConnectionStatusPending      RuntimeAgentConnectionStatus = iota
	RuntimeAgentConnectionStatusConnected    RuntimeAgentConnectionStatus = iota
	RuntimeAgentConnectionStatusDisconnected RuntimeAgentConnectionStatus = iota
)

type RuntimeConfig struct {
	KymaConfig    KymaConfig
	ClusterConfig interface{}
	Kubeconfig    string
}

type RuntimeStatus struct {
	LastOperationStatus     Operation
	RuntimeConnectionStatus RuntimeAgentConnectionStatus
	RuntimeConfiguration    RuntimeConfig
}

func (rc RuntimeConfig) GCPConfig() (GCPConfig, bool) {
	gcpConfig, ok := rc.ClusterConfig.(GCPConfig)

	return gcpConfig, ok
}

func (rc RuntimeConfig) GardenerConfig() (GardenerConfig, bool) {
	gardenerConfig, ok := rc.ClusterConfig.(GardenerConfig)

	return gardenerConfig, ok
}

func RuntimeConfigFromInput(input *gqlschema.ProvisionRuntimeInput) RuntimeConfig {
	return RuntimeConfig{
		KymaConfig:    kymaConfigFromInput(*input.KymaConfig),
		ClusterConfig: clusterConfigFromInput(*input.ClusterConfig),
	}
}

func RuntimeStatusToGraphQLStatus(status RuntimeStatus) *gqlschema.RuntimeStatus {
	return &gqlschema.RuntimeStatus{
		LastOperationStatus:     OperationStatusToGQLOperationStatus(status.LastOperationStatus),
		RuntimeConnectionStatus: runtimeConnectionStatusToGraphQLStatus(status.RuntimeConnectionStatus),
		RuntimeConfiguration:    runtimeConfigurationToGraphQLConfiguration(status.RuntimeConfiguration),
	}
}

func OperationStatusToGQLOperationStatus(operation Operation) *gqlschema.OperationStatus {
	return &gqlschema.OperationStatus{
		Operation: operationTypeToGraphQLType(operation.Operation),
		State:     operationStateToGraphQLState(operation.State),
		Message:   operation.Message,
		RuntimeID: operation.RuntimeID,
	}
}

func clusterConfigFromInput(input gqlschema.ClusterConfigInput) interface{} {
	if input.GardenerConfig != nil {
		config := input.GardenerConfig
		return gardenerConfigFromInput(*config)
	}
	if input.GcpConfig != nil {
		config := input.GcpConfig
		return gcpConfigFromInput(*config)
	}
	return nil
}

func gardenerConfigFromInput(input gqlschema.GardenerConfigInput) GardenerConfig {
	return GardenerConfig{
		Name:              input.Name,
		ProjectName:       input.ProjectName,
		KubernetesVersion: input.KubernetesVersion,
		NodeCount:         input.NodeCount,
		VolumeSize:        input.VolumeSize,
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
	}
}

func gcpConfigFromInput(input gqlschema.GCPConfigInput) GCPConfig {
	return GCPConfig{
		Name:              input.Name,
		ProjectName:       input.ProjectName,
		KubernetesVersion: input.KubernetesVersion,
		NumberOfNodes:     input.NumberOfNodes,
		BootDiskSize:      input.BootDiskSize,
		MachineType:       input.MachineType,
		Region:            input.Region,
		Zone:              *input.Zone,
	}
}

func kymaConfigFromInput(input gqlschema.KymaConfigInput) KymaConfig {
	var modules []KymaModule
	for _, module := range input.Modules {
		modules = append(modules, KymaModule(module))
	}

	return KymaConfig{
		Version: input.Version,
		Modules: modules,
	}
}

func runtimeConnectionStatusToGraphQLStatus(status RuntimeAgentConnectionStatus) *gqlschema.RuntimeConnectionStatus {
	return &gqlschema.RuntimeConnectionStatus{Status: runtimeAgentConnectionStatusToGraphQLStatus(status)}
}

func runtimeAgentConnectionStatusToGraphQLStatus(status RuntimeAgentConnectionStatus) gqlschema.RuntimeAgentConnectionStatus {
	switch status {
	case RuntimeAgentConnectionStatusConnected:
		return gqlschema.RuntimeAgentConnectionStatusConnected
	case RuntimeAgentConnectionStatusDisconnected:
		return gqlschema.RuntimeAgentConnectionStatusDisconnected
	case RuntimeAgentConnectionStatusPending:
		return gqlschema.RuntimeAgentConnectionStatusPending
	default:
		return ""
	}
}

func runtimeConfigurationToGraphQLConfiguration(config RuntimeConfig) *gqlschema.RuntimeConfig {
	return &gqlschema.RuntimeConfig{
		ClusterConfig: clusterConfigToGraphQLConfig(config.ClusterConfig),
		KymaConfig:    kymaConfigToGraphQLConfig(config.KymaConfig),
		Kubeconfig:    &config.Kubeconfig,
	}
}

func clusterConfigToGraphQLConfig(config interface{}) gqlschema.ClusterConfig {
	gardenerConfig, ok := config.(GardenerConfig)
	if ok {
		return gardenerConfigToGraphQLConfig(gardenerConfig)
	}

	gcpConfig, ok := config.(GCPConfig)
	if ok {
		return gcpConfigToGraphQLConfig(gcpConfig)
	}
	return nil
}

func gardenerConfigToGraphQLConfig(config GardenerConfig) gqlschema.ClusterConfig {
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

func gcpConfigToGraphQLConfig(config GCPConfig) gqlschema.ClusterConfig {
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

func kymaConfigToGraphQLConfig(config KymaConfig) *gqlschema.KymaConfig {
	var modules []*gqlschema.KymaModule
	for _, module := range config.Modules {
		kymaModule := gqlschema.KymaModule(module)
		modules = append(modules, &kymaModule)
	}

	return &gqlschema.KymaConfig{
		Version: &config.Version,
		Modules: modules,
	}
}

func operationTypeToGraphQLType(operationType OperationType) gqlschema.OperationType {
	switch operationType {
	case Provision:
		return gqlschema.OperationTypeProvision
	case Deprovision:
		return gqlschema.OperationTypeDeprovision
	case Upgrade:
		return gqlschema.OperationTypeUpgrade
	case ReconnectRuntime:
		return gqlschema.OperationTypeReconnectRuntime
	default:
		return ""
	}
}

func operationStateToGraphQLState(state OperationState) gqlschema.OperationState {
	switch state {
	case InProgress:
		return gqlschema.OperationStateInProgress
	case Succeeded:
		return gqlschema.OperationStateSucceeded
	case Failed:
		return gqlschema.OperationStateFailed
	default:
		return ""
	}
}
