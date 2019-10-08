package model

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"time"
)

type InfrastructureProvider int

const (
	GCP      InfrastructureProvider = iota
	AKS      InfrastructureProvider = iota
	Gardener InfrastructureProvider = iota
)

type KymaModule int

const (
	BackupModule             KymaModule = iota
	BackupInitModule         KymaModule = iota
	JaegerModule             KymaModule = iota
	LoggingModule            KymaModule = iota
	MonitoringModule         KymaModule = iota
	PrometheusOperatorModule KymaModule = iota
	KialiModule              KymaModule = iota
	KnativeBuildModule       KymaModule = iota
)

type KymaConfig struct {
	Version string
	Modules []KymaModule
}

type OperationState int

const (
	Pending    OperationState = iota
	InProgress OperationState = iota
	Succeeded  OperationState = iota
	Failed     OperationState = iota
)

type OperationType int

const (
	Provision        OperationType = iota
	Upgrade          OperationType = iota
	Deprovision      OperationType = iota
	ReconnectRuntime OperationType = iota
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

func (rc RuntimeConfig) GCPConfig() (GCPConfig, bool) {
	gcpConfig, ok := rc.ClusterConfig.(GCPConfig)

	return gcpConfig, ok
}

func (rc RuntimeConfig) GardenerConfig() (GardenerConfig, bool) {
	gardenerConfig, ok := rc.ClusterConfig.(GardenerConfig)

	return gardenerConfig, ok
}

func RuntimeConfigFromInput(input gqlschema.ProvisionRuntimeInput) RuntimeConfig {
	return RuntimeConfig{
		KymaConfig:    kymaConfigFromInput(*input.KymaConfig),
		ClusterConfig: clusterConfigFromInput(*input.ClusterConfig),
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
	for module := range input.Modules {
		modules = append(modules, KymaModule(module))
	}

	return KymaConfig{
		Version: input.Version,
		Modules: modules,
	}
}

type RuntimeStatus struct {
	LastOperationStatus     Operation
	RuntimeConnectionStatus RuntimeAgentConnectionStatus
	RuntimeConfiguration    RuntimeConfig
}

func OperationStatusToGQLOperationStatus(operation Operation) *gqlschema.OperationStatus {
	return &gqlschema.OperationStatus{
		Operation: gqlschema.OperationType(string(operation.Operation)),
		State:     gqlschema.OperationState(string(operation.State)),
		Message:   operation.Message,
		RuntimeID: operation.RuntimeID,
	}
}
