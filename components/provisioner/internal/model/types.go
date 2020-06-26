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
	WaitingForClusterDomain   OperationStage = "WaitingForClusterDomain"
	WaitingForClusterCreation OperationStage = "WaitingForClusterCreation"
	StartingInstallation      OperationStage = "StartingInstallation"
	WaitingForInstallation    OperationStage = "WaitingForInstallation"
	ConnectRuntimeAgent       OperationStage = "ConnectRuntimeAgent"
	WaitForAgentToConnect     OperationStage = "WaitForAgentToConnect"

	TriggerKymaUninstall   OperationStage = "TriggerKymaUninstall"
	WaitForClusterDeletion OperationStage = "WaitForClusterDeletion"
	DeleteCluster          OperationStage = "DeprovisionCluster"
	CleanupCluster         OperationStage = "CleanupCluster"

	StartingUpgrade      OperationStage = "StartingUpgrade"
	UpdatingUpgradeState OperationStage = "UpdatingUpgradeState"

	FinishedStage OperationStage = "Finished"
)

type Cluster struct {
	ID                 string
	Kubeconfig         *string
	CreationTimestamp  time.Time
	Deleted            bool
	Tenant             string
	SubAccountId       *string
	ActiveKymaConfigId string

	ClusterConfig GardenerConfig `db:"-"`
	KymaConfig    KymaConfig     `db:"-"`
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

type OperationsCount struct {
	Count map[OperationType]int
}
