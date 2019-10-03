package persistence

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"

	"github.com/gofrs/uuid"
)

type RuntimeService interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, error)
	SetProvisioningStarted(runtimeID string, clusterConfig model.ClusterConfig, kymaConfig model.KymaConfig) (model.Operation, error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, error)
	SetUpgradeStarted(runtimeID string) (model.Operation, error)
	GetLastOperation(runtimeID string) (model.Operation, error)
	Update(runtimeID string, kubeconfig string, deprovisioningContext string) error
}

type runtimeService struct {
	repository Repository
}

func NewRuntimeService(repository Repository) RuntimeService {
	return runtimeService{
		repository: repository,
	}
}

func (r runtimeService) GetStatus(runtimeID string) (model.RuntimeStatus, error) {
	return model.RuntimeStatus{}, nil
}

func (r runtimeService) SetProvisioningStarted(runtimeID string, clusterConfig model.ClusterConfig, kymaConfig model.KymaConfig) (model.Operation, error) {

	started := time.Now()

	transaction, err := r.repository.BeginTransaction()
	if err != nil {
		return model.Operation{}, err
	}

	err = r.repository.InsertCluster(runtimeID, started, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, err
	}

	err = r.repository.InsertClusterConfig(runtimeID, clusterConfig, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, err
	}

	err = r.repository.InsertKymaConfig(runtimeID, kymaConfig, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, err
	}

	operation := model.Operation{
		OperationID: uuid.String(),
		Operation:   model.Provision,
		Started:     started,
		State:       model.InProgress,
		RuntimeID:   runtimeID,
	}

	err = r.repository.InsertOperation(operation, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, err
	}

	err = transaction.Commit()
	if err != nil {
		logrus.Errorf("Failed to commit transaction for runtime: '%s'.", runtimeID)
		return model.Operation{}, err
	}

	return operation, nil
}

func rollback(transaction *sql.Tx, runtimeID string) {
	err := transaction.Rollback()
	if err != nil {
		logrus.Errorf("Failed to rollback transaction for runtime: '%s'.", runtimeID)
	}
}

func (r runtimeService) SetDeprovisioningStarted(runtimeID string) (model.Operation, error) {
	return model.Operation{}, nil
}

func (r runtimeService) SetUpgradeStarted(runtimeID string) (model.Operation, error) {
	return model.Operation{}, nil
}

func (r runtimeService) GetLastOperation(runtimeID string) (model.Operation, error) {
	return model.Operation{}, nil
}

func (r runtimeService) Update(runtimeID string, kubeconfig string, deprovisioningContext string) error {
	return nil
}
