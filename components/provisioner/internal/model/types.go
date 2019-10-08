package model

import (
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
	KubernetesVersion string
	NodeCount         int
	VolumeSize        string
	MachineType       string
	TargetProvider    string
	TargetSecret      string
	Cidr              string
	Region            string
	Zone              string
	AutoScalerMin     string
	AutoScalerMax     string
	MaxSurge          int
	MaxUnavailable    int
}

type GCPConfig struct {
	Name              string
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

type RuntimeStatus struct {
	LastOperationStatus     Operation
	RuntimeConnectionStatus RuntimeAgentConnectionStatus
	RuntimeConfiguration    RuntimeConfig
}
