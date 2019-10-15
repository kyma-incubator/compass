package persistence

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=OperationService
type RuntimeService interface {
	GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error)
	SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error)
	SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error)
	SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error)
	GetLastOperation(runtimeID string) (model.Operation, dberrors.Error)
	Update(runtimeID string, kubeconfig string, terraformState string) dberrors.Error
	CleanupData(runtimeID string) dberrors.Error
}

type runtimeService struct {
	dbSessionFactory dbsession.Factory
}

func NewRuntimeService(dbSessionFactory dbsession.Factory) RuntimeService {
	return runtimeService{
		dbSessionFactory: dbSessionFactory,
	}
}

func (r runtimeService) GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

	operation, err := session.GetLastOperation(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	clusterConfig, err := session.GetClusterConfig(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	kymaConfig, err := session.GetKymaConfig(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	runtimeConfiguration := model.RuntimeConfig{
		KymaConfig:    kymaConfig,
		ClusterConfig: clusterConfig,
	}

	return model.RuntimeStatus{
		LastOperationStatus:  operation,
		RuntimeConfiguration: runtimeConfiguration,
	}, nil
}

func (r runtimeService) SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error) {
	dbSession, err := r.dbSessionFactory.NewSessionWithinTransaction()
	if err != nil {
		logrus.Errorf("Failed to create repository: %s", err)
	}

	timestamp := time.Now()

	err = dbSession.InsertCluster(runtimeID, timestamp, "{}")
	if err != nil {
		rollback(dbSession, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	gcpConfig, isGCP := runtimeConfig.GCPConfig()
	if isGCP {
		err = dbSession.InsertGCPConfig(runtimeID, gcpConfig)
		if err != nil {
			rollback(dbSession, runtimeID)
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	gardenerConfig, isGardener := runtimeConfig.GardenerConfig()
	if isGardener {
		err = dbSession.InsertGardenerConfig(runtimeID, gardenerConfig)
		if err != nil {
			rollback(dbSession, runtimeID)
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	kymaConfigID, err := dbSession.InsertKymaConfig(runtimeID, runtimeConfig.KymaConfig.Version)
	if err != nil {
		rollback(dbSession, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	for _, module := range runtimeConfig.KymaConfig.Modules {
		err = dbSession.InsertKymaConfigModule(kymaConfigID, module)
		if err != nil {
			rollback(dbSession, runtimeID)
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	operation := model.Operation{
		Type:           model.Provision,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        "Provisioning started",
		ClusterID:      runtimeID,
	}

	operationID, err := dbSession.InsertOperation(operation)

	if err != nil {
		rollback(dbSession, runtimeID)
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	operation.ID = operationID

	err = dbSession.Commit()
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	return operation, nil
}

func (r runtimeService) SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return r.setOperationStarted(runtimeID, model.Deprovision, "Deprovisioning started.", "Deprovisioning failed: %s")
}

func (r runtimeService) SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error) {
	return r.setOperationStarted(runtimeID, model.Upgrade, "Upgrade started.", "Upgrade failed: %s")
}

func (r runtimeService) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

	return session.GetLastOperation(runtimeID)
}

func (r runtimeService) Update(runtimeID string, kubeconfig string, terraformState string) dberrors.Error {
	session := r.dbSessionFactory.NewWriteSession()

	return session.UpdateCluster(runtimeID, kubeconfig, terraformState)
}

func (r runtimeService) CleanupData(runtimeID string) dberrors.Error {
	session := r.dbSessionFactory.NewWriteSession()

	return session.DeleteCluster(runtimeID)
}

func (r runtimeService) setOperationStarted(runtimeID string, operationType model.OperationType, message string, errorMessageFmt string) (model.Operation, dberrors.Error) {
	dbSession := r.dbSessionFactory.NewWriteSession()

	operation := model.Operation{
		Type:           operationType,
		StartTimestamp: time.Now(),
		State:          model.InProgress,
		Message:        message,
		ClusterID:      runtimeID,
	}

	operationID, err := dbSession.InsertOperation(operation)
	if err != nil {
		return model.Operation{}, dberrors.Internal(errorMessageFmt, err)
	}

	operation.ID = operationID

	return operation, nil
}

func rollback(transaction dbsession.Transaction, runtimeID string) {
	err := transaction.Rollback()
	if err != nil {
		logrus.Errorf("Failed to rollback transaction for runtime: '%s'.", runtimeID)
	}
}
