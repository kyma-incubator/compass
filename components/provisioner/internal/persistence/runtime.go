package persistence

import (
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"log"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
)

type RuntimeService interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error)
	SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	Update(runtimeID string, kubeconfig string, deprovisioningContext string) dberrors.Error
}

type runtimeService struct {
	repositoryFactory RepositoryFactory
}

func NewRuntimeService(repositoryFactory RepositoryFactory) RuntimeService {
	return runtimeService{
		repositoryFactory: repositoryFactory,
	}
}

func (r runtimeService) GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	return model.RuntimeStatus{}, nil
}

func (r runtimeService) SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error) {

	repository, err := r.repositoryFactory.New()
	if err != nil {
		logrus.Errorf("Failed to create repository: %s", err)
	}

	timestamp := time.Now()
	transaction := repository.Transaction()

	err = repository.InsertCluster(runtimeID, timestamp, "{}")
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	gcpConfig, isGCP := runtimeConfig.GCPConfig()
	if isGCP {
		err = repository.InsertGCPConfig(runtimeID, gcpConfig)
		if err != nil {
			rollback(transaction, runtimeID)
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	gardenerConfig, isGardener := runtimeConfig.GardenerConfig()
	if isGardener {
		err = repository.InsertGardenerConfig(runtimeID, gardenerConfig)
		if err != nil {
			rollback(transaction, runtimeID)
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	kymaConfigID, err := repository.InsertKymaConfig(runtimeID, runtimeConfig.KymaConfig.Version)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	for _, module := range runtimeConfig.KymaConfig.Modules {
		err = repository.InsertKymaConfigModule(kymaConfigID, module)
		if err != nil {
			rollback(transaction, runtimeID)
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	id, e := uuid.NewUUID()
	if e != nil {
		log.Println("Failed to create UUID")
	}

	operation := model.Operation{
		OperationID: id.String(),
		Operation: model.Provision,
		Started: timestamp,
		State: model.InProgress,
		Message: "Provisioning started",
		RuntimeID: runtimeID,
	}

	err = repository.InsertOperation(operation)
	if err != nil {
		rollback(transaction, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	err = transaction.Commit()
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}


	return model.Operation{}, nil
}

func (r runtimeService) SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func (r runtimeService) SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func rollback(transaction Transaction, runtimeID string) {
	err := transaction.Rollback()
	if err != nil {
		logrus.Errorf("Failed to rollback transaction for runtime: '%s'.", runtimeID)
	}
}

func (r runtimeService) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	return model.Operation{}, nil
}

func (r runtimeService) Update(runtimeID string, kubeconfig string, deprovisioningContext string) dberrors.Error {
	return nil
}
