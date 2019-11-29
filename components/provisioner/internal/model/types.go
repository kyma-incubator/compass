package model

import (
	"time"
)

type KymaModule string

type KymaConfig struct {
	ID        string
	Release   Release
	Modules   []KymaConfigModule
	ClusterID string
}

type KymaConfigModule struct {
	ID           string
	Module       KymaModule
	KymaConfigID string
}

type Release struct {
	Id            string
	Version       string
	TillerYAML    string
	InstallerYAML string
}

type GithubRelease struct {
	Id         int     `json:"id"`
	Name       string  `json:"name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	Url  string `json:"browser_download_url"`
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
	ID                    string
	Kubeconfig            *string
	TerraformState        []byte
	CredentialsSecretName string
	CreationTimestamp     time.Time

	ClusterConfig ProviderConfiguration `db:"-"`
	KymaConfig    KymaConfig            `db:"-"`
}

func (c Cluster) GCPConfig() (GCPConfig, bool) {
	gcpConfig, ok := c.ClusterConfig.(GCPConfig)

	return gcpConfig, ok
}

func (c Cluster) GardenerConfig() (GardenerConfig, bool) {
	gardenerConfig, ok := c.ClusterConfig.(GardenerConfig)

	return gardenerConfig, ok
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

type RuntimeAgentConnectionStatus int

const (
	RuntimeAgentConnectionStatusPending      RuntimeAgentConnectionStatus = iota
	RuntimeAgentConnectionStatusConnected    RuntimeAgentConnectionStatus = iota
	RuntimeAgentConnectionStatusDisconnected RuntimeAgentConnectionStatus = iota
)

type RuntimeStatus struct {
	LastOperationStatus     Operation
	RuntimeConnectionStatus RuntimeAgentConnectionStatus
	RuntimeConfiguration    Cluster
}
