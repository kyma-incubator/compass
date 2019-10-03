package model

import "time"

type InfrastructureProvider int

const (
	GCP      InfrastructureProvider = iota
	AKS      InfrastructureProvider = iota
	Gardener InfrastructureProvider = iota
)

type ClusterConfig struct {
	Name                   string
	NodeCount              int
	DiskSize               string
	MachineType            string
	ComputeZone            string
	Version                string
	InfrastructureProvider InfrastructureProvider
	ProviderConfig         interface{}
}

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

type AdditionalProperties map[string]interface{}

type GardenerProviderConfig struct {
	TargetProvider       string
	TargetSecret         string
	AutoScalerMin        string
	AutoScalerMax        string
	MaxSurge             int
	MaxUnavailable       int
	AdditionalProperties AdditionalProperties
}

type RuntimeAgentConnectionStatus int

const (
	RuntimeAgentConnectionStatusPending      RuntimeAgentConnectionStatus = iota
	RuntimeAgentConnectionStatusConnected    RuntimeAgentConnectionStatus = iota
	RuntimeAgentConnectionStatusDisconnected RuntimeAgentConnectionStatus = iota
)

type RuntimeConnectionConfig struct {
	Kubeconfig string
}

type RuntimeConfig struct {
	KymaConfig    KymaConfig
	ClusterConfig ClusterConfig
}

type RuntimeStatus struct {
	lastOperationStatus     Operation
	runtimeConnectionStatus RuntimeAgentConnectionStatus
	runtimeConnectionConfig RuntimeConnectionConfig
	runtimeConfiguration    RuntimeConfig
}

type GCPProviderConfig struct {
	AdditionalProperties AdditionalProperties
}

type AKSProviderConfig struct {
	AdditionalProperties AdditionalProperties
}
