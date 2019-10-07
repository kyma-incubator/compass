package persistence

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"

	"github.com/gofrs/uuid"
)

type RuntimeService interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	SetProvisioningStarted(runtimeID string, clusterConfig model.ClusterConfig, kymaConfig model.KymaConfig) (model.Operation, dberrors.Error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error)
	SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	Update(runtimeID string, kubeconfig string, deprovisioningContext string) dberrors.Error
}

type runtimeService struct {
	repository Repository
}

func NewRuntimeService(repository Repository) RuntimeService {
	return runtimeService{
		repository: repository,
	}
}

func (r runtimeService) GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	return model.RuntimeStatus{}, nil
}

func (r runtimeService) SetProvisioningStarted(runtimeID string, clusterConfig model.ClusterConfig, kymaConfig model.KymaConfig) (model.Operation, dberrors.Error) {

	started := time.Now()

	transaction, err := r.repository.BeginTransaction()
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to start transaction: %s.", err)
	}

	err = r.repository.InsertCluster(runtimeID, started, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to insert data to Cluster table: %s.", err)
	}

	err = r.repository.InsertClusterConfig(runtimeID, clusterConfig, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to insert cluster config data: %s.", err)
	}

	err = r.repository.InsertKymaConfig(runtimeID, kymaConfig, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to insert Kyma config data: %s", err)
	}

	id, err := uuid.NewV4()
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to generate UUID: %s", err)
	}

	operation := model.Operation{
		OperationID: id.String(),
		Operation:   model.Provision,
		Started:     started,
		State:       model.InProgress,
		RuntimeID:   runtimeID,
	}

	err = r.repository.InsertOperation(operation, transaction)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to insert operation data: %s", err)
	}

	err = transaction.Commit()
	if err != nil {
		logrus.Errorf("Failed to commit transaction for runtime: '%s'.", runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to commit transaction: %s.", err)
	}

	return operation, nil
}

func rollback(transaction *sql.Tx, runtimeID string) {
	err := transaction.Rollback()
	if err != nil {
		logrus.Errorf("Failed to rollback transaction for runtime: '%s'.", runtimeID)
	}
}

func (r runtimeService) SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func (r runtimeService) SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func (r runtimeService) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func (r runtimeService) Update(runtimeID string, kubeconfig string, deprovisioningContext string) dberrors.Error {
	return nil
}
