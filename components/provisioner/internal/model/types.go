package model

import (
	"time"
)

type KymaModule string

type KymaConfig struct {
	ID        string
	Version   string
	Modules   []KymaConfigModule
	ClusterID string
}

type KymaConfigModule struct {
	ID           string
	Module       KymaModule
	KymaConfigID string
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

type Cluster struct {
	ID                string
	Kubeconfig        string
	TerraformState    string
	CreationTimestamp time.Time
}

type Operation struct {
	ID             string
	Type           OperationType
	StartTimestamp time.Time
	EndTimestamp   *time.Time
	State          OperationState
	Message        string
	ClusterID      string
}

type GardenerConfig struct {
	ID                string
	ClusterID         string
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
	ID                string
	ClusterID         string
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
	ID             string
	ClusterID      string
	Name           string
	NodeCount      int
	DiskSize       string
	MachineType    string
	Region         string
	Version        string
	Credentials    string
	ProviderConfig interface{}
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
