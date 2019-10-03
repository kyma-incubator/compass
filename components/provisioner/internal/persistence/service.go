package persistence

import "github.com/kyma-incubator/compass/components/provisioner/internal/model"

type OperationService interface {
	Get(operationID string) (model.Operation, error)
	SetAsFailed(operationID string, message string) error
	SetAsSucceeded(operationID string) error
}

type RuntimeService interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, error)
	SetProvisioningStarted(runtimeID string, clusterConfig model.ClusterConfig, kymaConfig model.KymaConfig) (model.Operation, error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, error)
	SetUpgradeStarted(runtimeID string) (model.Operation, error)
}
