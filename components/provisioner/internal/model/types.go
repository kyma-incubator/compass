package model

import (
	"time"
)

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

type OperationStage string

const (
	ShootProvisioning      OperationStage = "ShootProvisioning"
	StartingInstallation   OperationStage = "StartingInstallation"
	WaitingForInstallation OperationStage = "WaitingForInstallation"
	ConnectRuntimeAgent    OperationStage = "ConnectRuntimeAgent"
	WaitForAgentToConnect  OperationStage = "WaitForAgentToConnect"

	StartingUpgrade      OperationStage = "StartingUpgrade"
	UpdatingUpgradeState OperationStage = "UpdatingUpgradeState"

	FinishedStage OperationStage = "Finished"

	DeprovisioningStage OperationStage = "Deprovisioning"
)

type Cluster struct {
	ID             string
	Kubeconfig     *string
	TerraformState []byte
	// CredentialsSecretName is ignored for now, as only Gardener implementation is supported
	CredentialsSecretName string
	CreationTimestamp     time.Time
	Deleted               bool
	Tenant                string
	SubAccountId          string
	ActiveKymaConfigId    string

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
	Stage          OperationStage
	LastTransition *time.Time
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
